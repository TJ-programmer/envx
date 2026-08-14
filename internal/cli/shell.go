package cli

import (
	"errors"
	"fmt"
	"io"

	"envx/internal/errs"
)

func cmdShell(args []string, stdout, stderr io.Writer) int {
	root, rest := splitRootFlag(args)
	if checkHelp(rest, stdout, "shell") {
		return 0
	}
	envName := ""
	shellCmd := ""
	overlay := false
	for i := 0; i < len(rest); i++ {
		switch rest[i] {
		case "--env":
			if i+1 >= len(rest) {
				return printError(stderr, errors.New("missing value for '--env'"))
			}
			envName = rest[i+1]
			i++
		case "--shell":
			if i+1 >= len(rest) {
				return printError(stderr, errors.New("missing value for '--shell'"))
			}
			shellCmd = rest[i+1]
			i++
		case "--overlay":
			overlay = true
		default:
			return printError(stderr, fmt.Errorf("unknown option '%s'", rest[i]))
		}
	}

	service := projectService(root)
	if !service.IsInitialized() {
		return printError(stderr, errs.ErrProjectNotInitialized)
	}
	fmt.Fprintln(stderr, "Entering envx shell with secrets loaded (type 'exit' to return).")
	code, err := service.RunShell(shellCmd, envName, overlay)
	if err != nil {
		return printError(stderr, err)
	}
	return code
}
