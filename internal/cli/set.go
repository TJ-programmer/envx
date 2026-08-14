package cli

import (
	"errors"
	"fmt"
	"io"
)

func cmdSet(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) < 1 {
		return printError(stderr, errors.New("usage: envx set KEY [VALUE] [--env ENV] [--secret|--plain]"))
	}
	root, rest := splitRootFlag(args)
	if checkHelp(rest, stdout, "set") {
		return 0
	}
	key := rest[0]
	rest = rest[1:]
	envName := ""
	secret, plain := false, false
	value := ""
	hasValue := false
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
			if hasValue {
				return printError(stderr, fmt.Errorf("unexpected argument '%s'", rest[0]))
			}
			value = rest[0]
			hasValue = true
			rest = rest[1:]
		}
	}
	if secret && plain {
		return printError(stderr, errors.New("use either '--secret' or '--plain', not both"))
	}

	if !hasValue {
		if !secret {
			return printError(stderr, errors.New("missing value for KEY (or pass --secret to prompt securely)"))
		}
		prompted, err := readSecret(fmt.Sprintf("Secret for %s: ", key), stdin, stderr)
		if err != nil {
			return printError(stderr, err)
		}
		if prompted == "" {
			return printError(stderr, errors.New("no value provided"))
		}
		value = prompted
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
