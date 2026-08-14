package cli

import (
	"errors"
	"fmt"
	"io"
	"strconv"
)

func cmdConfig(args []string, stdout, stderr io.Writer) int {
	root, rest := splitRootFlag(args)
	if checkHelp(rest, stdout, "config") {
		return 0
	}
	if len(rest) < 1 {
		return printError(stderr, errors.New("usage: envx config get|set <key> [value]"))
	}
	service := projectService(root)

	switch rest[0] {
	case "get":
		if len(rest) != 2 {
			return printError(stderr, errors.New("usage: envx config get KEY"))
		}
		value, err := service.GetSetting(rest[1])
		if err != nil {
			return printError(stderr, err)
		}
		fmt.Fprintln(stdout, value)
		return 0
	case "set":
		if len(rest) != 3 {
			return printError(stderr, errors.New("usage: envx config set KEY VALUE"))
		}
		key, value := rest[1], rest[2]
		switch key {
		case "encryption.default_encrypt":
			parsed, err := strconv.ParseBool(value)
			if err != nil {
				return printError(stderr, fmt.Errorf("'%s' expects true or false", key))
			}
			if err := service.SetDefaultEncrypt(parsed); err != nil {
				return printError(stderr, err)
			}
			fmt.Fprintf(stdout, "Set %s=%t.\n", key, parsed)
			return 0
		case "migration.overlay_dotenv":
			parsed, err := strconv.ParseBool(value)
			if err != nil {
				return printError(stderr, fmt.Errorf("'%s' expects true or false", key))
			}
			if err := service.SetOverlayDotenv(parsed); err != nil {
				return printError(stderr, err)
			}
			fmt.Fprintf(stdout, "Set %s=%t.\n", key, parsed)
			return 0
		default:
			return printError(stderr, fmt.Errorf("unknown setting '%s'", key))
		}
	default:
		return printError(stderr, fmt.Errorf("unknown config subcommand '%s'", rest[0]))
	}
}
