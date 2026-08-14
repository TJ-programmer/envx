package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunOverlayLegacyDotenv(t *testing.T) {
	if os.Getenv("ENVX_TEST_HELPER") == "1" {
		if err := os.WriteFile(os.Getenv("ENVX_TEST_OUT"), []byte(os.Getenv("LEGACY_VAR")+"|"+os.Getenv("PORT")), 0o644); err != nil {
			os.Exit(1)
		}
		return
	}

	dir := t.TempDir()
	runIn(dir, "init")
	runIn(dir, "set", "PORT", "8123")
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("LEGACY_VAR=from-dotenv\nPORT=9999\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	outFile := filepath.Join(dir, "overlay-output.txt")
	t.Setenv("ENVX_TEST_HELPER", "1")
	t.Setenv("ENVX_TEST_OUT", outFile)

	code, _, errOut := runIn(dir, "run", "--overlay", "--", os.Args[0], "-test.run", "TestRunOverlayLegacyDotenv")
	if code != 0 {
		t.Fatalf("run --overlay failed (%d): %s", code, errOut)
	}
	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "from-dotenv|8123" {
		t.Fatalf("overlay result = %q, want from-dotenv|8123 (envx value must win)", data)
	}
}

func TestRunOverlayViaConfigSetting(t *testing.T) {
	if os.Getenv("ENVX_TEST_HELPER") == "1" {
		if err := os.WriteFile(os.Getenv("ENVX_TEST_OUT"), []byte(os.Getenv("LEGACY_VAR")), 0o644); err != nil {
			os.Exit(1)
		}
		return
	}

	dir := t.TempDir()
	runIn(dir, "init")
	code, _, errOut := runIn(dir, "config", "set", "migration.overlay_dotenv", "true")
	if code != 0 {
		t.Fatalf("config set failed (%d): %s", code, errOut)
	}
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("LEGACY_VAR=auto-overlaid\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	outFile := filepath.Join(dir, "overlay-setting-output.txt")
	t.Setenv("ENVX_TEST_HELPER", "1")
	t.Setenv("ENVX_TEST_OUT", outFile)

	code, _, errOut = runIn(dir, "run", "--", os.Args[0], "-test.run", "TestRunOverlayViaConfigSetting")
	if code != 0 {
		t.Fatalf("run failed (%d): %s", code, errOut)
	}
	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "auto-overlaid" {
		t.Fatalf("config-driven overlay result = %q", data)
	}
}

func TestRunNoOverlayWithoutFlag(t *testing.T) {
	if os.Getenv("ENVX_TEST_HELPER") == "1" {
		if err := os.WriteFile(os.Getenv("ENVX_TEST_OUT"), []byte(os.Getenv("LEGACY_VAR")), 0o644); err != nil {
			os.Exit(1)
		}
		return
	}

	dir := t.TempDir()
	runIn(dir, "init")
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("LEGACY_VAR=should-not-appear\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	outFile := filepath.Join(dir, "no-overlay-output.txt")
	t.Setenv("ENVX_TEST_HELPER", "1")
	t.Setenv("ENVX_TEST_OUT", outFile)

	code, _, errOut := runIn(dir, "run", "--", os.Args[0], "-test.run", "TestRunNoOverlayWithoutFlag")
	if code != 0 {
		t.Fatalf("run failed (%d): %s", code, errOut)
	}
	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "" {
		t.Fatalf("without --overlay .env should not be loaded, got %q", data)
	}
}

func TestKeyRotatePreservesSecrets(t *testing.T) {
	dir := t.TempDir()
	runIn(dir, "init")
	runIn(dir, "set", "DB_PASSWORD", "hunter2", "--secret")
	runIn(dir, "set", "PORT", "8000")

	oldKey, err := os.ReadFile(filepath.Join(dir, ".envx", "key.bin"))
	if err != nil {
		t.Fatal(err)
	}

	code, out, errOut := runIn(dir, "key", "rotate")
	if code != 0 {
		t.Fatalf("rotate failed (%d): %s", code, errOut)
	}
	if !strings.Contains(out, "1 secret") {
		t.Fatalf("rotate output should mention 1 re-encrypted secret: %s", out)
	}

	newKey, err := os.ReadFile(filepath.Join(dir, ".envx", "key.bin"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(oldKey, newKey) {
		t.Fatal("key.bin should change after rotate")
	}

	oldKeyBackup, err := os.ReadFile(filepath.Join(dir, ".envx", "key.old.bin"))
	if err != nil {
		t.Fatalf("key.old.bin backup should exist: %v", err)
	}
	if !bytes.Equal(oldKey, oldKeyBackup) {
		t.Fatal("key.old.bin should contain the previous key")
	}

	code, shown, _ := runIn(dir, "list", "--show-secrets")
	if code != 0 || !strings.Contains(shown, "hunter2") {
		t.Fatalf("secret not preserved after rotate:\n%s", shown)
	}
	code, value, _ := runIn(dir, "get", "DB_PASSWORD", "--show-secret")
	if code != 0 || strings.TrimSpace(value) != "hunter2" {
		t.Fatalf("get after rotate = %q", value)
	}
}

func TestKeyStatusAndExportImport(t *testing.T) {
	dir := t.TempDir()
	runIn(dir, "init")

	code, out, _ := runIn(dir, "key", "status")
	if code != 0 || !strings.Contains(out, "Present: true") {
		t.Fatalf("key status output:\n%s", out)
	}

	code, keyOut, errOut := runIn(dir, "key", "export")
	if code != 0 {
		t.Fatalf("key export failed (%d): %s", code, errOut)
	}
	key := strings.TrimSpace(keyOut)
	if len(key) != 44 {
		t.Fatalf("exported key length = %d, want 44", len(key))
	}

	keyFile := filepath.Join(dir, "exported.key")
	if err := os.WriteFile(keyFile, []byte(key+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, _, errOut = runIn(dir, "key", "import", keyFile)
	if code != 0 {
		t.Fatalf("key import failed (%d): %s", code, errOut)
	}
	imported, err := os.ReadFile(filepath.Join(dir, ".envx", "key.bin"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(imported)) != key {
		t.Fatalf("imported key mismatch")
	}

	garbage := filepath.Join(dir, "garbage.key")
	if err := os.WriteFile(garbage, []byte("not-a-key\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, _, errOut = runIn(dir, "key", "import", garbage)
	if code != 1 {
		t.Fatalf("importing an invalid key should fail, got %d", code)
	}
	if !strings.Contains(errOut, "invalid key") {
		t.Fatalf("stderr should mention invalid key: %s", errOut)
	}
}

func TestSetSecretPromptReadsStdin(t *testing.T) {
	dir := t.TempDir()
	runIn(dir, "init")

	old, _ := os.Getwd()
	defer os.Chdir(old)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := RunWithIO([]string{"set", "API_KEY", "--secret"}, strings.NewReader("typed-secret\n"), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("set --secret prompt failed (%d): %s", code, stderr.String())
	}

	code, value, _ := runIn(dir, "get", "API_KEY", "--show-secret")
	if code != 0 || strings.TrimSpace(value) != "typed-secret" {
		t.Fatalf("prompted secret not stored, got %q", value)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".envx", "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "enc:") {
		t.Fatalf("prompted secret should be stored encrypted:\n%s", data)
	}
}

func TestSetWithoutValueOrSecretFails(t *testing.T) {
	dir := t.TempDir()
	runIn(dir, "init")

	code, _, errOut := runIn(dir, "set", "API_KEY")
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if !strings.Contains(errOut, "--secret") {
		t.Fatalf("stderr should suggest --secret: %s", errOut)
	}
}
