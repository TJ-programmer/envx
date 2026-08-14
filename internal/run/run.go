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
	childEnv := append([]string{}, os.Environ()...)
	for key, value := range envVars {
		childEnv = append(childEnv, key+"="+value)
	}

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
