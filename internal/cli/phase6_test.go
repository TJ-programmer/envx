package cli

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"envx/internal/clipboard"
)

func TestCopyCommandWritesClipboard(t *testing.T) {
	dir := t.TempDir()
	runIn(dir, "init")
	runIn(dir, "set", "PORT", "8080")
	runIn(dir, "set", "API_KEY", "hunter2", "--secret")

	var captured string
	old := clipboard.Write
	clipboard.Write = func(content string) error { captured = content; return nil }
	t.Cleanup(func() { clipboard.Write = old })

	code, out, errOut := runIn(dir, "copy", "PORT")
	if code != 0 {
		t.Fatalf("copy failed (%d): %s", code, errOut)
	}
	if captured != "PORT=8080\n" {
		t.Errorf("captured = %q, want %q", captured, "PORT=8080\n")
	}
	if !strings.Contains(out, "Copied 'PORT'") {
		t.Errorf("out: %s", out)
	}

	code, _, errOut = runIn(dir, "copy", "API_KEY")
	if code != 0 {
		t.Fatalf("copy secret failed (%d): %s", code, errOut)
	}
	if captured != "API_KEY=hunter2\n" {
		t.Errorf("secret should be decrypted into clipboard, got %q", captured)
	}

	code, _, _ = runIn(dir, "copy", "PORT", "API_KEY")
	if code != 0 {
		t.Fatal("copy multiple failed")
	}
	if captured != "PORT=8080\nAPI_KEY=hunter2\n" {
		t.Errorf("multi captured = %q", captured)
	}

	code, _, errOut = runIn(dir, "copy", "MISSING")
	if code != 0 {
		t.Fatal("copy missing should not fail the command")
	}
	if !strings.Contains(errOut, "not found") {
		t.Errorf("errOut should mention missing secret: %s", errOut)
	}
	if captured != "PORT=8080\nAPI_KEY=hunter2\n" {
		t.Errorf("clipboard should be unchanged after missing key")
	}
}

func TestShellCommandInjectsEnvWithActiveMarker(t *testing.T) {
	if os.Getenv("ENVX_SHELL_HELPER") == "1" {
		content := os.Getenv("ENVX_ACTIVE") + "|" + os.Getenv("PORT")
		if err := os.WriteFile(os.Getenv("ENVX_SHELL_HELPER_OUT"), []byte(content), 0o644); err != nil {
			os.Exit(1)
		}
		return
	}

	dir := t.TempDir()
	runIn(dir, "init")
	runIn(dir, "set", "PORT", "9099")

	outFile := filepath.Join(dir, "shell-out.txt")
	t.Setenv("ENVX_SHELL_HELPER", "1")
	t.Setenv("ENVX_SHELL_HELPER_OUT", outFile)

	shellCmd := strconv.Quote(os.Args[0]) + " -test.run TestShellCommandInjectsEnvWithActiveMarker"
	code, _, errOut := runIn(dir, "shell", "--shell", shellCmd)
	if code != 0 {
		t.Fatalf("shell failed (%d): %s", code, errOut)
	}
	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "1|9099" {
		t.Errorf("got %q, want %q", string(data), "1|9099")
	}
}

func TestShellNotInitialized(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := runIn(dir, "shell")
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if !strings.Contains(errOut, "not initialized") {
		t.Errorf("errOut: %s", errOut)
	}
}
