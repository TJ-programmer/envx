package cli

import (
	"errors"
	"fmt"
	"io"
)

func cmdSet(args []string, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		return printError(stderr, errors.New("usage: envx set KEY VALUE [--env ENV] [--secret|--plain]"))
	}
	root, rest := splitRootFlag(args)
	key, value := rest[0], rest[1]
	rest = rest[2:]
	envName := ""
	secret, plain := false, false
	for len(rest) > 0 {
		switch rest[0] {
		case "--env":
			if len(rest) < 2 {
				return printError(stderr, errors.New("missing value for '--env'"))
			}
			envName = rest[1]
			rest = rest[2:]
		case "--secret":
			secret = true
			rest = rest[1:]
		case "--plain":
			plain = true
			rest = rest[1:]
		default:
			return printError(stderr, fmt.Errorf("unknown option '%s'", rest[0]))
		}
	}
	if secret && plain {
		return printError(stderr, errors.New("use either '--secret' or '--plain', not both"))
	}

	service := projectService(root)
	cfg, err := service.SetVariable(key, value, envName, secret, plain)
	if err != nil {
		return printError(stderr, err)
	}
	env := envName
	if env == "" {
		env = cfg.ActiveEnv
	}
	mode := "plain"
	if cfg.Environments[env].Variables[key].IsSecret {
		mode = "secret"
	}
	fmt.Fprintf(stdout, "Stored '%s' in environment '%s' (%s).\n", key, env, mode)
	return 0
}
