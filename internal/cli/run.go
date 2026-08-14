package cli

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"envx/internal/config"
)

const watchPollInterval = 500 * time.Millisecond

func cmdRun(args []string, stdout, stderr io.Writer) int {
	root, args := splitRootFlag(args)
	if checkHelp(args, stdout, "run") {
		return 0
	}
	envName := ""
	shellCmd := ""
	overlay := false
	watch := false
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
		case "--watch":
			watch = true
		default:
			return printError(stderr, fmt.Errorf("unknown option '%s'", flags[i]))
		}
	}

	service := projectService(root)
	if watch {
		paths := config.ResolveProjectPaths(resolveRoot(root))
		watchPaths := []string{
			filepath.Join(paths.Root, ".env"),
			paths.ConfigPath,
			paths.KeyPath,
		}
		fmt.Fprintln(stderr, "[envx] watch mode enabled: restarting on .env/config changes (Ctrl+C to stop)")
		code, err := service.RunWatch(cmdArgs, shellCmd, envName, overlay, watchPollInterval, stderr, watchPaths)
		if err != nil {
			return printError(stderr, err)
		}
		return code
	}

	code, overlayCount, err := service.RunCommand(cmdArgs, shellCmd, envName, overlay)
	if err != nil {
		return printError(stderr, err)
	}
	if overlayCount > 0 {
		fmt.Fprintf(stderr, "Overlay: merged %d variable(s) from legacy .env\n", overlayCount)
	}
	return code
}
