package run

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

const helperVar = "ENVX_RUN_HELPER"

func runHelper() {
	content := os.Getenv("ENVX_ACTIVE") + "|" + os.Getenv("PORT")
	if err := os.WriteFile(os.Getenv("ENVX_RUN_HELPER_OUT"), []byte(content), 0o644); err != nil {
		os.Exit(1)
	}
}

func TestRunInjectsEnvxActive(t *testing.T) {
	if os.Getenv(helperVar) == "1" {
		runHelper()
		return
	}
	outFile := filepath.Join(t.TempDir(), "out.txt")
	t.Setenv(helperVar, "1")
	t.Setenv("ENVX_RUN_HELPER_OUT", outFile)

	r := NewRunner()
	code, err := r.Run([]string{os.Args[0], "-test.run", "TestRunInjectsEnvxActive"}, "", map[string]string{"PORT": "8123"})
	if err != nil || code != 0 {
		t.Fatalf("run failed: code=%d err=%v", code, err)
	}
	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "1|8123" {
		t.Errorf("got %q, want %q", string(data), "1|8123")
	}
}

func TestStartShellInjectsEnvxActive(t *testing.T) {
	if os.Getenv(helperVar) == "1" {
		runHelper()
		return
	}
	outFile := filepath.Join(t.TempDir(), "out.txt")
	t.Setenv(helperVar, "1")
	t.Setenv("ENVX_RUN_HELPER_OUT", outFile)

	r := NewRunner()
	shellCmd := strconv.Quote(os.Args[0]) + " -test.run TestStartShellInjectsEnvxActive"
	proc, err := r.StartShell(shellCmd, map[string]string{"PORT": "9090"})
	if err != nil {
		t.Fatal(err)
	}
	code, err := proc.Wait()
	if err != nil || code != 0 {
		t.Fatalf("shell wait failed: code=%d err=%v", code, err)
	}
	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "1|9090" {
		t.Errorf("got %q, want %q", string(data), "1|9090")
	}
}
