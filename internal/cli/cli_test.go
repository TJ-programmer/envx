package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func runIn(dir string, args ...string) (int, string, string) {
	old, _ := os.Getwd()
	defer os.Chdir(old)
	if err := os.Chdir(dir); err != nil {
		return 1, "", err.Error()
	}
	var stdout, stderr bytes.Buffer
	code := Run(args, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func TestInitAndSetSecret(t *testing.T) {
	dir := t.TempDir()

	code, out, errOut := runIn(dir, "init")
	if code != 0 {
		t.Fatalf("init failed (%d): %s", code, errOut)
	}
	_ = out

	code, out, errOut = runIn(dir, "set", "API_KEY", "my-secret", "--secret")
	if code != 0 {
		t.Fatalf("set failed (%d): %s", code, errOut)
	}
	_ = out

	data, err := os.ReadFile(filepath.Join(dir, ".envx", "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}
	environments := cfg["environments"].(map[string]any)
	dev := environments["dev"].(map[string]any)
	variables := dev["variables"].(map[string]any)
	apiKey := variables["API_KEY"].(map[string]any)
	stored := apiKey["value"].(string)
	if !strings.HasPrefix(stored, "enc:") {
		t.Fatalf("stored value %q does not start with enc:", stored)
	}
}

func TestListRedactsByDefaultAndCanShowSecrets(t *testing.T) {
	dir := t.TempDir()
	runIn(dir, "init")
	runIn(dir, "set", "API_KEY", "my-secret", "--secret")

	listedCode, listed, _ := runIn(dir, "list")
	shownCode, shown, _ := runIn(dir, "list", "--show-secrets")

	if listedCode != 0 || shownCode != 0 {
		t.Fatal("list commands failed")
	}
	if !strings.Contains(listed, "********") {
		t.Errorf("listed output should redact secrets:\n%s", listed)
	}
	if strings.Contains(listed, "my-secret") {
		t.Errorf("listed output should not contain the secret:\n%s", listed)
	}
	if !strings.Contains(shown, "my-secret") {
		t.Errorf("show-secrets output should reveal the secret:\n%s", shown)
	}
}

func TestRunInjectsEnvironmentVectorMode(t *testing.T) {
	if os.Getenv("ENVX_TEST_HELPER") == "1" {
		if err := os.WriteFile(os.Getenv("ENVX_TEST_OUT"), []byte(os.Getenv("PORT")), 0o644); err != nil {
			os.Exit(1)
		}
		return
	}

	dir := t.TempDir()
	runIn(dir, "init")
	runIn(dir, "set", "PORT", "8123")

	outFile := filepath.Join(dir, "vector-output.txt")
	t.Setenv("ENVX_TEST_HELPER", "1")
	t.Setenv("ENVX_TEST_OUT", outFile)

	code, _, errOut := runIn(dir, "run", "--", os.Args[0], "-test.run", "TestRunInjectsEnvironmentVectorMode")
	if code != 0 {
		t.Fatalf("run failed (%d): %s", code, errOut)
	}
	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "8123" {
		t.Errorf("injected PORT = %q, want 8123", data)
	}
}

func TestRunInjectsEnvironmentShellMode(t *testing.T) {
	if os.Getenv("ENVX_TEST_HELPER") == "1" {
		if err := os.WriteFile(os.Getenv("ENVX_TEST_OUT"), []byte(os.Getenv("SHELL_SECRET")), 0o644); err != nil {
			os.Exit(1)
		}
		return
	}

	dir := t.TempDir()
	runIn(dir, "init")
	runIn(dir, "set", "SHELL_SECRET", "works")

	outFile := filepath.Join(dir, "shell-output.txt")
	t.Setenv("ENVX_TEST_HELPER", "1")
	t.Setenv("ENVX_TEST_OUT", outFile)

	shellCmd := strconv.Quote(os.Args[0]) + " -test.run TestRunInjectsEnvironmentShellMode"
	code, _, errOut := runIn(dir, "run", "--shell", shellCmd)
	if code != 0 {
		t.Fatalf("run failed (%d): %s", code, errOut)
	}
	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "works" {
		t.Errorf("injected SHELL_SECRET = %q, want works", data)
	}
}

func TestCorruptConfigReturnsClearError(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".envx"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".envx", "config.json"), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}

	code, _, errOut := runIn(dir, "list")
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if !strings.Contains(errOut, "invalid JSON") {
		t.Fatalf("stderr should mention invalid JSON: %s", errOut)
	}
}

func TestMissingKeyFailsSafely(t *testing.T) {
	dir := t.TempDir()
	runIn(dir, "init")
	runIn(dir, "set", "API_KEY", "my-secret", "--secret")
	if err := os.Remove(filepath.Join(dir, ".envx", "key.bin")); err != nil {
		t.Fatal(err)
	}

	code, _, errOut := runIn(dir, "list", "--show-secrets")
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if !strings.Contains(strings.ToLower(errOut), "key not found") {
		t.Fatalf("stderr should mention the missing key: %s", errOut)
	}
}

func TestEnvironmentCommandsWork(t *testing.T) {
	dir := t.TempDir()
	runIn(dir, "init")

	createCode, _, _ := runIn(dir, "env", "create", "staging")
	useCode, _, _ := runIn(dir, "env", "use", "staging")
	listCode, listed, _ := runIn(dir, "env", "list")

	if createCode != 0 || useCode != 0 || listCode != 0 {
		t.Fatal("env commands failed")
	}
	if !strings.Contains(listed, "staging") {
		t.Errorf("env list should contain staging:\n%s", listed)
	}
	if !strings.Contains(listed, "true") {
		t.Errorf("env list should mark staging active:\n%s", listed)
	}
}

func TestInitIsIdempotentOnlyWithForce(t *testing.T) {
	dir := t.TempDir()
	first, _, _ := runIn(dir, "init")
	second, _, errOut := runIn(dir, "init")
	forced, _, _ := runIn(dir, "init", "--force")

	if first != 0 {
		t.Fatal("first init failed")
	}
	if second != 1 {
		t.Fatalf("second init should fail without --force, got %d", second)
	}
	if !strings.Contains(errOut, "--force") {
		t.Fatalf("stderr should suggest --force: %s", errOut)
	}
	if forced != 0 {
		t.Fatal("init --force should succeed")
	}
}

func TestUnknownCommandFails(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := runIn(dir, "bogus")
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if !strings.Contains(errOut, "unknown command") {
		t.Fatalf("stderr: %s", errOut)
	}
}

func TestListJSONFormat(t *testing.T) {
	dir := t.TempDir()
	runIn(dir, "init")
	runIn(dir, "set", "PORT", "8000")

	code, out, _ := runIn(dir, "list", "--format", "json")
	if code != 0 {
		t.Fatal("list --format json failed")
	}
	var rows []map[string]string
	if err := json.Unmarshal([]byte(out), &rows); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out)
	}
	if len(rows) != 1 || rows[0]["key"] != "PORT" {
		t.Errorf("rows = %v", rows)
	}
}
