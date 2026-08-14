package cli

import (
	"errors"
	"fmt"
	"io"
)

func cmdExport(args []string, stdout, stderr io.Writer) int {
	root, rest := splitRootFlag(args)
	envName := ""
	format := "shell"
	for len(rest) > 0 {
		switch rest[0] {
		case "--env":
			if len(rest) < 2 {
				return printError(stderr, errors.New("missing value for '--env'"))
			}
			envName = rest[1]
			rest = rest[2:]
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
	out, err := service.ExportEnv(envName, format)
	if err != nil {
		return printError(stderr, err)
	}
	fmt.Fprint(stdout, out)
	return 0
}
