package cli

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"envx/internal/clipboard"
	"envx/internal/errs"
)

func cmdCopy(args []string, stdout, stderr io.Writer) int {
	root, rest := splitRootFlag(args)
	if checkHelp(rest, stdout, "copy") {
		return 0
	}
	envName := ""
	keys := []string{}
	for len(rest) > 0 {
		switch rest[0] {
		case "--env":
			if len(rest) < 2 {
				return printError(stderr, errors.New("missing value for '--env'"))
			}
			envName = rest[1]
			rest = rest[2:]
		default:
			if strings.HasPrefix(rest[0], "-") {
				return printError(stderr, fmt.Errorf("unknown option '%s'", rest[0]))
			}
			keys = append(keys, rest[0])
			rest = rest[1:]
		}
	}
	if len(keys) == 0 {
		return printError(stderr, errors.New("usage: envx copy KEY [KEY...] [--env ENV] [--root DIR]"))
	}

	service := projectService(root)
	var content strings.Builder
	copied := 0
	notFound := []string{}
	for _, key := range keys {
		value, err := service.GetVariable(key, envName, true)
		if err != nil {
			if errors.Is(err, errs.ErrVariableNotFound) {
				notFound = append(notFound, key)
				continue
			}
			return printError(stderr, err)
		}
		fmt.Fprintf(&content, "%s=%s\n", key, value)
		copied++
	}

	if copied > 0 {
		if err := clipboard.Write(content.String()); err != nil {
			return printError(stderr, err)
		}
		if copied == 1 {
			fmt.Fprintf(stdout, "Copied '%s' to clipboard.\n", keys[0])
		} else {
			fmt.Fprintf(stdout, "Copied %d secret(s) to clipboard.\n", copied)
		}
	}
	for _, key := range notFound {
		fmt.Fprintf(stderr, "Secret '%s' not found.\n", key)
	}
	return 0
}
