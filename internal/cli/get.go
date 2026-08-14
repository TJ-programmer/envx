package cli

import (
	"errors"
	"fmt"
	"io"
)

func cmdGet(args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 {
		return printError(stderr, errors.New("usage: envx get KEY [--env ENV] [--show-secret]"))
	}
	root, rest := splitRootFlag(args)
	key := rest[0]
	rest = rest[1:]
	envName := ""
	showSecret := false
	for len(rest) > 0 {
		switch rest[0] {
		case "--env":
			if len(rest) < 2 {
				return printError(stderr, errors.New("missing value for '--env'"))
			}
			envName = rest[1]
			rest = rest[2:]
		case "--show-secret":
			showSecret = true
			rest = rest[1:]
		default:
			return printError(stderr, fmt.Errorf("unknown option '%s'", rest[0]))
		}
	}

	service := projectService(root)
	value, err := service.GetVariable(key, envName, showSecret)
	if err != nil {
		return printError(stderr, err)
	}
	fmt.Fprintln(stdout, value)
	return 0
}
