package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetShowsValueAndRedactsSecrets(t *testing.T) {
	dir := t.TempDir()
	runIn(dir, "init")
	runIn(dir, "set", "PORT", "8080")
	runIn(dir, "set", "DB_PASSWORD", "hunter2", "--secret")

	code, out, errOut := runIn(dir, "get", "PORT")
	if code != 0 {
		t.Fatalf("get failed (%d): %s", code, errOut)
	}
	if strings.TrimSpace(out) != "8080" {
		t.Fatalf("get PORT = %q, want 8080", out)
	}

	code, out, _ = runIn(dir, "get", "DB_PASSWORD")
	if code != 0 {
		t.Fatal("get secret failed")
	}
	if strings.TrimSpace(out) != "********" {
		t.Fatalf("get DB_PASSWORD should redact by default, got %q", out)
	}

	code, out, _ = runIn(dir, "get", "DB_PASSWORD", "--show-secret")
	if code != 0 {
		t.Fatal("get --show-secret failed")
	}
	if strings.TrimSpace(out) != "hunter2" {
		t.Fatalf("get --show-secret = %q, want hunter2", out)
	}
}

func TestGetMissingVariableFails(t *testing.T) {
	dir := t.TempDir()
	runIn(dir, "init")

	code, _, errOut := runIn(dir, "get", "NOPE")
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if !strings.Contains(errOut, "not found") {
		t.Fatalf("stderr should mention not found: %s", errOut)
	}
}

func TestUnsetRemovesVariable(t *testing.T) {
	dir := t.TempDir()
	runIn(dir, "init")
	runIn(dir, "set", "PORT", "8080")

	code, _, errOut := runIn(dir, "unset", "PORT")
	if code != 0 {
		t.Fatalf("unset failed (%d): %s", code, errOut)
	}

	code, out, _ := runIn(dir, "list")
	if code != 0 {
		t.Fatal("list failed")
	}
	if strings.Contains(out, "PORT") {
		t.Fatalf("list should not contain PORT after unset:\n%s", out)
	}

	code, _, errOut = runIn(dir, "unset", "PORT")
	if code != 1 {
		t.Fatalf("unsetting a missing variable should fail, got %d", code)
	}
}

func TestImportDotenvFile(t *testing.T) {
	dir := t.TempDir()
	runIn(dir, "init")

	envFile := filepath.Join(dir, "app.env")
	content := "# comment\nPORT=9000\nDB_HOST='localhost'\nexport APP_NAME=\"demo app\"\nAPI_TOKEN=shh\n"
	if err := os.WriteFile(envFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	code, out, errOut := runIn(dir, "import", envFile)
	if code != 0 {
		t.Fatalf("import failed (%d): %s", code, errOut)
	}
	if !strings.Contains(out, "4") {
		t.Fatalf("import should report 4 variables, got: %s", out)
	}

	code, plain, _ := runIn(dir, "get", "DB_HOST")
	if code != 0 || strings.TrimSpace(plain) != "localhost" {
		t.Fatalf("DB_HOST not imported correctly: %q", plain)
	}
	code, quoted, _ := runIn(dir, "get", "APP_NAME")
	if code != 0 || strings.TrimSpace(quoted) != "demo app" {
		t.Fatalf("APP_NAME not imported correctly: %q", quoted)
	}

	code, token, _ := runIn(dir, "get", "API_TOKEN", "--show-secret")
	if code != 0 || strings.TrimSpace(token) != "shh" {
		t.Fatalf("API_TOKEN not imported: %q", token)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".envx", "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "enc:") {
		t.Fatalf("sensitive-named API_TOKEN should be stored encrypted:\n%s", data)
	}
}

func TestImportMissingFileFails(t *testing.T) {
	dir := t.TempDir()
	runIn(dir, "init")

	code, _, errOut := runIn(dir, "import", filepath.Join(dir, "nope.env"))
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if !strings.Contains(errOut, "nope.env") {
		t.Fatalf("stderr should mention the missing file: %s", errOut)
	}
}

func TestExportFormats(t *testing.T) {
	dir := t.TempDir()
	runIn(dir, "init")
	runIn(dir, "set", "PORT", "8000")
	runIn(dir, "set", "GREETING", "hello world")

	code, shell, errOut := runIn(dir, "export")
	if code != 0 {
		t.Fatalf("export failed (%d): %s", code, errOut)
	}
	if !strings.Contains(shell, "export PORT='8000'") {
		t.Fatalf("shell export missing PORT: %s", shell)
	}
	if !strings.Contains(shell, "export GREETING='hello world'") {
		t.Fatalf("shell export missing quoted GREETING: %s", shell)
	}

	code, dotenv, _ := runIn(dir, "export", "--format", "dotenv")
	if code != 0 {
		t.Fatal("export dotenv failed")
	}
	if !strings.Contains(dotenv, "PORT=8000") {
		t.Fatalf("dotenv export missing PORT: %s", dotenv)
	}
	if !strings.Contains(dotenv, `GREETING="hello world"`) {
		t.Fatalf("dotenv export should quote spaces: %s", dotenv)
	}

	code, jsonOut, _ := runIn(dir, "export", "--format", "json")
	if code != 0 {
		t.Fatal("export json failed")
	}
	var payload map[string]string
	if err := json.Unmarshal([]byte(jsonOut), &payload); err != nil {
		t.Fatalf("json export not valid JSON: %v\n%s", err, jsonOut)
	}
	if payload["PORT"] != "8000" {
		t.Fatalf("json export = %v", payload)
	}

	code, _, errOut = runIn(dir, "export", "--format", "yaml")
	if code != 1 {
		t.Fatalf("unknown format should fail, got %d", code)
	}
}

func TestDiffEnvironments(t *testing.T) {
	dir := t.TempDir()
	runIn(dir, "init")
	runIn(dir, "set", "PORT", "8000")
	runIn(dir, "env", "create", "staging")
	runIn(dir, "set", "PORT", "9000", "--env", "staging")
	runIn(dir, "set", "FEATURE", "on", "--env", "staging")

	code, out, errOut := runIn(dir, "diff", "dev", "staging")
	if code != 0 {
		t.Fatalf("diff failed (%d): %s", code, errOut)
	}
	if !strings.Contains(out, "PORT") || !strings.Contains(out, "8000") || !strings.Contains(out, "9000") {
		t.Fatalf("diff should show PORT difference:\n%s", out)
	}
	if !strings.Contains(out, "FEATURE") || !strings.Contains(out, "-") {
		t.Fatalf("diff should show FEATURE only in staging:\n%s", out)
	}

	code, out, _ = runIn(dir, "diff", "dev", "dev")
	if code != 0 {
		t.Fatal("diff of same env failed")
	}
	if !strings.Contains(out, "No differences") {
		t.Fatalf("diff of identical envs should say no differences:\n%s", out)
	}
}

func TestConfigSettingToggle(t *testing.T) {
	dir := t.TempDir()
	runIn(dir, "init")

	code, out, _ := runIn(dir, "config", "get", "encryption.default_encrypt")
	if code != 0 || strings.TrimSpace(out) != "false" {
		t.Fatalf("default should be false, got %q", out)
	}

	code, _, errOut := runIn(dir, "config", "set", "encryption.default_encrypt", "true")
	if code != 0 {
		t.Fatalf("config set failed (%d): %s", code, errOut)
	}

	code, out, _ = runIn(dir, "config", "get", "encryption.default_encrypt")
	if code != 0 || strings.TrimSpace(out) != "true" {
		t.Fatalf("default_encrypt should now be true, got %q", out)
	}

	runIn(dir, "set", "PORT", "8000")
	data, err := os.ReadFile(filepath.Join(dir, ".envx", "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "enc:") {
		t.Fatalf("with default_encrypt on, plain sets should be encrypted:\n%s", data)
	}
}

func TestInitManagesGitignore(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	code, _, errOut := runIn(dir, "init")
	if code != 0 {
		t.Fatalf("init failed (%d): %s", code, errOut)
	}
	data, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("init should create .gitignore: %v", err)
	}
	if !strings.Contains(string(data), ".envx/") {
		t.Fatalf(".gitignore should ignore .envx/:\n%s", data)
	}

	runIn(dir, "init", "--force")
	data, err = os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(data), ".envx/") != 1 {
		t.Fatalf("gitignore should be deduped:\n%s", data)
	}
}

func TestInitNoGitignoreOptOut(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	code, _, errOut := runIn(dir, "init", "--no-gitignore")
	if code != 0 {
		t.Fatalf("init --no-gitignore failed (%d): %s", code, errOut)
	}
	if _, err := os.Stat(filepath.Join(dir, ".gitignore")); !os.IsNotExist(err) {
		t.Fatalf(".gitignore should not be created with --no-gitignore")
	}
}

func TestInitPreservesExistingGitignore(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	original := "# my rules\nnode_modules/\n"
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	code, _, errOut := runIn(dir, "init")
	if code != 0 {
		t.Fatalf("init failed (%d): %s", code, errOut)
	}
	data, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "node_modules/") {
		t.Fatalf("existing gitignore content lost:\n%s", content)
	}
	if !strings.Contains(content, "# envx") {
		t.Fatalf("envx section missing:\n%s", content)
	}
}

func TestRootDiscoveryFromSubdirectory(t *testing.T) {
	dir := t.TempDir()
	runIn(dir, "init")
	runIn(dir, "set", "PORT", "7777")

	sub := filepath.Join(dir, "nested", "deeper")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	code, out, errOut := runIn(sub, "get", "PORT")
	if code != 0 {
		t.Fatalf("get from subdir failed (%d): %s", code, errOut)
	}
	if strings.TrimSpace(out) != "7777" {
		t.Fatalf("get from subdir = %q, want 7777", out)
	}
}

func TestRootFlagOverridesDiscovery(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()
	runIn(dirA, "init")
	runIn(dirA, "set", "PORT", "1111")
	runIn(dirB, "init")
	runIn(dirB, "set", "PORT", "2222")

	code, out, errOut := runIn(dirB, "get", "PORT", "--root", dirA)
	if code != 0 {
		t.Fatalf("get --root failed (%d): %s", code, errOut)
	}
	if strings.TrimSpace(out) != "1111" {
		t.Fatalf("get --root = %q, want 1111", out)
	}
}

func TestDoctorReportsOk(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	runIn(dir, "init")
	runIn(dir, "set", "PORT", "8000")

	code, out, errOut := runIn(dir, "doctor")
	if code != 0 {
		t.Fatalf("doctor failed (%d): %s", code, errOut)
	}
	for _, want := range []string{"[ok]", "All checks passed"} {
		if !strings.Contains(out, want) {
			t.Fatalf("doctor output missing %q:\n%s", want, out)
		}
	}
}

func TestDoctorWarnsOnSensitivePlaintext(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	runIn(dir, "init")
	runIn(dir, "set", "DB_PASSWORD", "oops")

	code, out, _ := runIn(dir, "doctor")
	if code != 1 {
		t.Fatalf("doctor should flag warnings, got %d", code)
	}
	if !strings.Contains(out, "[warn]") || !strings.Contains(out, "DB_PASSWORD") {
		t.Fatalf("doctor should warn about sensitive plaintext:\n%s", out)
	}
}

func TestDoctorNotInitialized(t *testing.T) {
	dir := t.TempDir()
	code, out, _ := runIn(dir, "doctor")
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if !strings.Contains(out, "not initialized") {
		t.Fatalf("doctor output:\n%s", out)
	}
}

func TestCompletionsGenerate(t *testing.T) {
	dir := t.TempDir()
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		code, out, errOut := runIn(dir, "completions", shell)
		if code != 0 {
			t.Fatalf("completions %s failed (%d): %s", shell, code, errOut)
		}
		if out == "" {
			t.Fatalf("completions %s produced no output", shell)
		}
	}
	code, _, _ := runIn(dir, "completions", "tcsh")
	if code != 1 {
		t.Fatalf("unknown shell should fail, got %d", code)
	}
}
