package run

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"envx/internal/errs"
)

type Runner struct{}

func NewRunner() *Runner {
	return &Runner{}
}

type Process interface {
	Wait() (int, error)
	Kill() error
}

type childProcess struct {
	cmd *exec.Cmd
}

func (r *Runner) Run(argv []string, shellCmd string, envVars map[string]string) (int, error) {
	proc, err := r.Start(argv, shellCmd, envVars)
	if err != nil {
		return 1, err
	}
	return proc.Wait()
}

func (r *Runner) Start(argv []string, shellCmd string, envVars map[string]string) (Process, error) {
	childEnv := buildChildEnv(envVars)

	var cmd *exec.Cmd
	if shellCmd != "" {
		cmd = buildShellCmd(shellCmd)
	} else {
		if len(argv) == 0 {
			return nil, fmt.Errorf("%w: no command was provided. Use 'envx run -- <command>'.", errs.ErrCommandExecution)
		}
		cmd = exec.Command(argv[0], argv[1:]...)
	}

	cmd.Env = childEnv
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("%w: failed to start child process: %v", errs.ErrCommandExecution, err)
	}
	return &childProcess{cmd: cmd}, nil
}

func (p *childProcess) Wait() (int, error) {
	err := p.cmd.Wait()
	if err == nil {
		return 0, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode(), nil
	}
	return 1, fmt.Errorf("%w: failed to execute child process: %v", errs.ErrCommandExecution, err)
}

func (p *childProcess) Kill() error {
	if p.cmd.Process == nil {
		return nil
	}
	return p.cmd.Process.Kill()
}

// StartShell launches an interactive shell with the resolved variables loaded.
// When shellCmd is non-empty it is executed as a one-shot command string (like
// run --shell); otherwise the platform's default interactive shell is started
// with a prompt indicator. The child environment always includes ENVX_ACTIVE=1.
func (r *Runner) StartShell(shellCmd string, envVars map[string]string) (Process, error) {
	childEnv := buildChildEnv(envVars)

	var cmd *exec.Cmd
	var cleanup func()
	if shellCmd != "" {
		cmd = buildShellCmd(shellCmd)
	} else {
		name, args, extraEnv, c := interactiveShell()
		cmd = exec.Command(name, args...)
		childEnv = append(childEnv, extraEnv...)
		cleanup = c
	}

	cmd.Env = childEnv
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("%w: failed to start interactive shell: %v", errs.ErrCommandExecution, err)
	}
	return &cleanupProcess{Process: &childProcess{cmd: cmd}, cleanup: cleanup}, nil
}

func buildChildEnv(envVars map[string]string) []string {
	childEnv := append([]string{}, os.Environ()...)
	childEnv = append(childEnv, "ENVX_ACTIVE=1")
	for key, value := range envVars {
		childEnv = append(childEnv, key+"="+value)
	}
	return childEnv
}

type cleanupProcess struct {
	Process
	cleanup func()
}

func (p *cleanupProcess) Wait() (int, error) {
	code, err := p.Process.Wait()
	if p.cleanup != nil {
		p.cleanup()
	}
	return code, err
}
