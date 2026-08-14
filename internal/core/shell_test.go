package core_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"envx/internal/config"
	"envx/internal/core"
	"envx/internal/crypto"
	"envx/internal/errs"
	"envx/internal/keyring"
	"envx/internal/run"
	"envx/internal/store"
)

type capturingRunner struct {
	captured map[string]string
	shellCmd string
}

func (c *capturingRunner) Run(argv []string, shellCmd string, envVars map[string]string) (int, error) {
	return 0, nil
}

func (c *capturingRunner) Start(argv []string, shellCmd string, envVars map[string]string) (run.Process, error) {
	return nil, nil
}

func (c *capturingRunner) StartShell(shellCmd string, envVars map[string]string) (run.Process, error) {
	c.captured = envVars
	c.shellCmd = shellCmd
	return &fakeProcess{}, nil
}

type fakeProcess struct{}

func (f *fakeProcess) Wait() (int, error) { return 0, nil }
func (f *fakeProcess) Kill() error        { return nil }

func shellService(t *testing.T, runner *capturingRunner) *core.EnvxService {
	t.Helper()
	paths := config.ResolveProjectPaths(t.TempDir())
	return core.NewService(
		store.New(paths),
		crypto.NewCryptoManager(keyring.NewLocalKeyProvider(paths)),
		runner,
	)
}

func TestRunShellResolvesEnvironment(t *testing.T) {
	runner := &capturingRunner{}
	service := shellService(t, runner)
	if _, err := service.InitProject("dev", false); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SetVariable("PORT", "8080", "", false, false); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SetVariable("API_KEY", "shh", "", true, false); err != nil {
		t.Fatal(err)
	}

	code, err := service.RunShell("", "", false)
	if err != nil || code != 0 {
		t.Fatalf("RunShell failed: code=%d err=%v", code, err)
	}
	if runner.captured["PORT"] != "8080" {
		t.Errorf("PORT not injected: %v", runner.captured)
	}
	if runner.captured["API_KEY"] != "shh" {
		t.Errorf("API_KEY should be decrypted for shell: %v", runner.captured)
	}
}

func TestRunShellAppliesOverlay(t *testing.T) {
	runner := &capturingRunner{}
	service := shellService(t, runner)
	if _, err := service.InitProject("dev", false); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SetVariable("PORT", "8080", "", false, false); err != nil {
		t.Fatal(err)
	}

	dotenv := filepath.Join(service.StoreRoot(), ".env")
	if err := os.WriteFile(dotenv, []byte("EXTRA=from-dotenv\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := service.RunShell("", "", true); err != nil {
		t.Fatal(err)
	}
	if runner.captured["EXTRA"] != "from-dotenv" {
		t.Errorf("overlay variable not injected: %v", runner.captured)
	}

	if _, err := service.RunShell("", "", false); err != nil {
		t.Fatal(err)
	}
	if _, ok := runner.captured["EXTRA"]; ok {
		t.Errorf("overlay variable should not be injected without overlay: %v", runner.captured)
	}
}

func TestRunShellForwardsShellCommand(t *testing.T) {
	runner := &capturingRunner{}
	service := shellService(t, runner)
	if _, err := service.InitProject("dev", false); err != nil {
		t.Fatal(err)
	}
	if _, err := service.RunShell("echo hi", "", false); err != nil {
		t.Fatal(err)
	}
	if runner.shellCmd != "echo hi" {
		t.Errorf("shellCmd = %q, want %q", runner.shellCmd, "echo hi")
	}
}

func TestRunShellRequiresInitializedProject(t *testing.T) {
	runner := &capturingRunner{}
	service := shellService(t, runner)
	code, err := service.RunShell("", "", false)
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if !strings.Contains(err.Error(), errs.ErrProjectNotInitialized.Error()) {
		t.Errorf("err = %v", err)
	}
}
