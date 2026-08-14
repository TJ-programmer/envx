package cli

import (
	"errors"
	"fmt"
	"io"
)

func cmdImport(args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 {
		return printError(stderr, errors.New("usage: envx import FILE [--env ENV]"))
	}
	root, rest := splitRootFlag(args)
	if checkHelp(rest, stdout, "import") {
		return 0
	}
	path := rest[0]
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
	count, err := service.ImportEnvFile(path, envName)
	if err != nil {
		return printError(stderr, err)
	}
	fmt.Fprintf(stdout, "Imported %d variables.\n", count)
	return 0
}
