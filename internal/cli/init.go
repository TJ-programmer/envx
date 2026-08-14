package cli

import (
	"errors"
	"fmt"
	"io"
	"os"

	"envx/internal/gitignore"
	"envx/internal/keyring"
)

func cmdInit(args []string, stdout, stderr io.Writer) int {
	root, rest := splitRootFlag(args)
	if checkHelp(rest, stdout, "init") {
		return 0
	}
	envName := "dev"
	force := false
	skipGitignore := false
	backend := ""
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
		case "--backend":
			if len(rest) < 2 {
				return printError(stderr, errors.New("missing value for '--backend'"))
			}
			backend = rest[1]
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

	if backend != "" {
		if backend != "file" && backend != "keyring" {
			return printError(stderr, fmt.Errorf("unknown key backend '%s' (use 'file' or 'keyring')", backend))
		}
		if backend == "keyring" && !keyring.Supported() {
			return printError(stderr, errors.New("the 'keyring' backend is not supported on this platform (Windows only)"))
		}
		if err := service.SetKeyBackend(backend); err != nil {
			return printError(stderr, err)
		}
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
