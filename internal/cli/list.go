package cli

import (
	"errors"
	"fmt"
	"io"
)

func cmdList(args []string, stdout, stderr io.Writer) int {
	root, rest := splitRootFlag(args)
	if checkHelp(rest, stdout, "list") {
		return 0
	}
	envName := ""
	showSecrets := false
	format := "table"
	for len(rest) > 0 {
		switch rest[0] {
		case "--env":
			if len(rest) < 2 {
				return printError(stderr, errors.New("missing value for '--env'"))
			}
			envName = rest[1]
			rest = rest[2:]
		case "--show-secrets":
			showSecrets = true
			rest = rest[1:]
		case "--format":
			if len(rest) < 2 {
				return printError(stderr, errors.New("missing value for '--format'"))
			}
			format = rest[1]
			rest = rest[2:]
		default:
			return printError(stderr, fmt.Errorf("unknown option '%s'", rest[0]))
		}
	}

	service := projectService(root)
	rows, err := service.ListVariables(envName, showSecrets)
	if err != nil {
		return printError(stderr, err)
	}
	renderVariableRows(rows, format, stdout)
	return 0
}
