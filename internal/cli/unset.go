package cli

import (
	"errors"
	"fmt"
	"io"
)

func cmdUnset(args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 {
		return printError(stderr, errors.New("usage: envx unset KEY [--env ENV]"))
	}
	root, rest := splitRootFlag(args)
	if checkHelp(rest, stdout, "unset") {
		return 0
	}
	key := rest[0]
	rest = rest[1:]
	envName := ""
	for len(rest) > 0 {
		switch rest[0] {
		case "--env":
			if len(rest) < 2 {
				return printError(stderr, errors.New("missing value for '--env'"))
			}
			envName = rest[1]
			rest = rest[2:]
		default:
			return printError(stderr, fmt.Errorf("unknown option '%s'", rest[0]))
		}
	}

	service := projectService(root)
	if err := service.UnsetVariable(key, envName); err != nil {
		return printError(stderr, err)
	}
	fmt.Fprintf(stdout, "Removed '%s'.\n", key)
	return 0
}
