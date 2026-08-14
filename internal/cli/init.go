package cli

import (
	"errors"
	"fmt"
	"io"
	"os"

	"envx/internal/gitignore"
)

func cmdInit(args []string, stdout, stderr io.Writer) int {
	root, rest := splitRootFlag(args)
	envName := "dev"
	force := false
	skipGitignore := false
	for len(rest) > 0 {
		switch rest[0] {
		case "--force":
			force = true
			rest = rest[1:]
		case "--env":
			if len(rest) < 2 {
				return printError(stderr, errors.New("missing value for '--env'"))
			}
			envName = rest[1]
			rest = rest[2:]
		case "--no-gitignore":
			skipGitignore = true
			rest = rest[1:]
		default:
			return printError(stderr, fmt.Errorf("unknown option '%s'", rest[0]))
		}
	}

	service := cwdService(root)
	cfg, err := service.InitProject(envName, force)
	if err != nil {
		return printError(stderr, err)
	}

	base := root
	if base == "" {
		base, _ = os.Getwd()
	}
	if err := gitignore.Ensure(base, skipGitignore); err != nil {
		return printError(stderr, err)
	}

	if skipGitignore {
		fmt.Fprintf(stdout, "Initialized envx in '.envx/' with active environment '%s' (gitignore skipped).\n", cfg.ActiveEnv)
	} else {
		fmt.Fprintf(stdout, "Initialized envx in '.envx/' with active environment '%s'.\n", cfg.ActiveEnv)
	}
	return 0
}
