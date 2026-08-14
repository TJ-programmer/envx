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

func (r *Runner) Run(argv []string, shellCmd string, envVars map[string]string) (int, error) {
	childEnv := append([]string{}, os.Environ()...)
	for key, value := range envVars {
		childEnv = append(childEnv, key+"="+value)
	}

	var cmd *exec.Cmd
	if shellCmd != "" {
		cmd = buildShellCmd(shellCmd)
	} else {
		if len(argv) == 0 {
			return 1, fmt.Errorf("%w: no command was provided. Use 'envx run -- <command>'.", errs.ErrCommandExecution)
		}
		cmd = exec.Command(argv[0], argv[1:]...)
	}

	cmd.Env = childEnv
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err == nil {
		return 0, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode(), nil
	}
	return 1, fmt.Errorf("%w: failed to execute child process: %v", errs.ErrCommandExecution, err)
}
