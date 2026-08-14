package cli

import (
	"errors"
	"io"
)

func cmdDiff(args []string, stdout, stderr io.Writer) int {
	root, rest := splitRootFlag(args)
	if len(rest) != 2 {
		return printError(stderr, errors.New("usage: envx diff ENV_A ENV_B"))
	}
	envA, envB := rest[0], rest[1]

	service := projectService(root)
	rows, err := service.DiffEnvironments(envA, envB)
	if err != nil {
		return printError(stderr, err)
	}
	renderDiffRows(rows, envA, envB, stdout)
	return 0
}
