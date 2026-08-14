package cli

import (
	"errors"
	"fmt"
	"io"
)

func cmdEnv(args []string, stdout, stderr io.Writer) int {
	root, rest := splitRootFlag(args)
	if checkHelp(rest, stdout, "env") {
		return 0
	}
	if len(rest) == 0 {
		return printError(stderr, errors.New("usage: envx env create|use|delete|list [NAME]"))
	}
	sub := rest[0]
	rest = rest[1:]
	service := projectService(root)

	switch sub {
	case "create":
		if len(rest) != 1 {
			return printError(stderr, errors.New("usage: envx env create NAME"))
		}
		if err := service.CreateEnvironment(rest[0]); err != nil {
			return printError(stderr, err)
		}
		fmt.Fprintf(stdout, "Created environment '%s'.\n", rest[0])
		return 0
	case "use":
		if len(rest) != 1 {
			return printError(stderr, errors.New("usage: envx env use NAME"))
		}
		if err := service.UseEnvironment(rest[0]); err != nil {
			return printError(stderr, err)
		}
		fmt.Fprintf(stdout, "Active environment is now '%s'.\n", rest[0])
		return 0
	case "delete":
		if len(rest) != 1 {
			return printError(stderr, errors.New("usage: envx env delete NAME"))
		}
		if err := service.DeleteEnvironment(rest[0]); err != nil {
			return printError(stderr, err)
		}
		fmt.Fprintf(stdout, "Deleted environment '%s'.\n", rest[0])
		return 0
	case "list":
		rows, err := service.ListEnvironments()
		if err != nil {
			return printError(stderr, err)
		}
		renderEnvironmentRows(rows, stdout)
		return 0
	default:
		return printError(stderr, fmt.Errorf("unknown env subcommand '%s'", sub))
	}
}
