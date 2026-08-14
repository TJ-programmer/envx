package cli

import (
	"errors"
	"fmt"
	"io"
)

func cmdRun(args []string, stdout, stderr io.Writer) int {
	root, args := splitRootFlag(args)
	envName := ""
	shellCmd := ""
	overlay := false
	cmdArgs := []string{}

	separator := -1
	for i, arg := range args {
		if arg == "--" {
			separator = i
			break
		}
	}
	flags := args
	if separator >= 0 {
		flags = args[:separator]
		cmdArgs = args[separator+1:]
	}
	for i := 0; i < len(flags); i++ {
		switch flags[i] {
		case "--env":
			if i+1 >= len(flags) {
				return printError(stderr, errors.New("missing value for '--env'"))
			}
			envName = flags[i+1]
			i++
		case "--shell":
			if i+1 >= len(flags) {
				return printError(stderr, errors.New("missing value for '--shell'"))
			}
			shellCmd = flags[i+1]
			i++
		case "--overlay":
			overlay = true
		default:
			return printError(stderr, fmt.Errorf("unknown option '%s'", flags[i]))
		}
	}

	service := projectService(root)
	code, overlayCount, err := service.RunCommand(cmdArgs, shellCmd, envName, overlay)
	if err != nil {
		return printError(stderr, err)
	}
	if overlayCount > 0 {
		fmt.Fprintf(stderr, "Overlay: merged %d variable(s) from legacy .env\n", overlayCount)
	}
	return code
}
