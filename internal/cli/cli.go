package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"envx/internal/bootstrap"
	"envx/internal/core"
)

func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stdout)
		return 0
	}

	switch args[0] {
	case "init":
		return cmdInit(args[1:], stdout, stderr)
	case "set":
		return cmdSet(args[1:], stdout, stderr)
	case "list":
		return cmdList(args[1:], stdout, stderr)
	case "run":
		return cmdRun(args[1:], stdout, stderr)
	case "env":
		return cmdEnv(args[1:], stdout, stderr)
	case "help", "-h", "--help":
		printUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "Error: unknown command '%s'\n", args[0])
		printUsage(stderr)
		return 1
	}
}

func cmdInit(args []string, stdout, stderr io.Writer) int {
	envName := "dev"
	force := false
	rest := args
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
		default:
			return printError(stderr, fmt.Errorf("unknown option '%s'", rest[0]))
		}
	}

	service := bootstrap.BuildService("")
	cfg, err := service.InitProject(envName, force)
	if err != nil {
		return printError(stderr, err)
	}
	fmt.Fprintf(stdout, "Initialized envx in '.envx/' with active environment '%s'.\n", cfg.ActiveEnv)
	return 0
}

func cmdSet(args []string, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		return printError(stderr, errors.New("usage: envx set KEY VALUE [--env ENV] [--secret|--plain]"))
	}
	key, value := args[0], args[1]
	rest := args[2:]
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

	service := bootstrap.BuildService("")
	cfg, err := service.SetVariable(key, value, envName, secret, plain)
	if err != nil {
		return printError(stderr, err)
	}
	env := envName
	if env == "" {
		env = cfg.ActiveEnv
	}
	fmt.Fprintf(stdout, "Stored '%s' in environment '%s'.\n", key, env)
	return 0
}

func cmdList(args []string, stdout, stderr io.Writer) int {
	envName := ""
	showSecrets := false
	format := "table"
	rest := args
	for len(rest) > 0 {
		switch rest[0] {
		case "--env":
			if len(rest) < 2 {
				return printError(stderr, errors.New("missing value for '--env'"))
			}
			envName = rest[1]
			rest = rest[2:]
		case "--show-secrets":
			showSecrets = true
			rest = rest[1:]
		case "--format":
			if len(rest) < 2 {
				return printError(stderr, errors.New("missing value for '--format'"))
			}
			format = rest[1]
			rest = rest[2:]
		default:
			return printError(stderr, fmt.Errorf("unknown option '%s'", rest[0]))
		}
	}

	service := bootstrap.BuildService("")
	rows, err := service.ListVariables(envName, showSecrets)
	if err != nil {
		return printError(stderr, err)
	}
	renderVariableRows(rows, format, stdout)
	return 0
}

func cmdRun(args []string, stdout, stderr io.Writer) int {
	envName := ""
	shellCmd := ""
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
		default:
			return printError(stderr, fmt.Errorf("unknown option '%s'", flags[i]))
		}
	}

	service := bootstrap.BuildService("")
	code, err := service.RunCommand(cmdArgs, shellCmd, envName)
	if err != nil {
		return printError(stderr, err)
	}
	return code
}

func cmdEnv(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		return printError(stderr, errors.New("usage: envx env create|use|delete|list [NAME]"))
	}
	sub := args[0]
	rest := args[1:]
	service := bootstrap.BuildService("")

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

func printError(stderr io.Writer, err error) int {
	fmt.Fprintf(stderr, "Error: %v\n", err)
	return 1
}

func renderVariableRows(rows []core.VariableRow, format string, stdout io.Writer) {
	if format == "json" {
		payload := make([]map[string]string, 0, len(rows))
		for _, row := range rows {
			payload = append(payload, map[string]string{
				"key":         row.Key,
				"value":       row.Value,
				"secret":      row.Secret,
				"environment": row.Environment,
			})
		}
		data, _ := json.MarshalIndent(payload, "", "  ")
		fmt.Fprintln(stdout, string(data))
		return
	}

	if len(rows) == 0 {
		fmt.Fprintln(stdout, "No values found.")
		return
	}

	widths := []int{len("KEY"), len("VALUE"), len("SECRET"), len("ENVIRONMENT")}
	for _, row := range rows {
		widths[0] = maxInt(widths[0], len(row.Key))
		widths[1] = maxInt(widths[1], len(row.Value))
		widths[2] = maxInt(widths[2], len(row.Secret))
		widths[3] = maxInt(widths[3], len(row.Environment))
	}
	fmt.Fprintf(stdout, "%-*s  %-*s  %-*s  %-*s\n", widths[0], "KEY", widths[1], "VALUE", widths[2], "SECRET", widths[3], "ENVIRONMENT")
	for _, row := range rows {
		fmt.Fprintf(stdout, "%-*s  %-*s  %-*s  %-*s\n", widths[0], row.Key, widths[1], row.Value, widths[2], row.Secret, widths[3], row.Environment)
	}
}

func renderEnvironmentRows(rows []core.EnvironmentRow, stdout io.Writer) {
	widths := []int{len("NAME"), len("ACTIVE"), len("VARIABLES")}
	for _, row := range rows {
		widths[0] = maxInt(widths[0], len(row.Name))
		widths[1] = maxInt(widths[1], len(row.Active))
		widths[2] = maxInt(widths[2], len(row.Variables))
	}
	fmt.Fprintf(stdout, "%-*s  %-*s  %-*s\n", widths[0], "NAME", widths[1], "ACTIVE", widths[2], "VARIABLES")
	for _, row := range rows {
		fmt.Fprintf(stdout, "%-*s  %-*s  %-*s\n", widths[0], row.Name, widths[1], row.Active, widths[2], row.Variables)
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "envx - project-local environment variable manager")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  envx <command> [options]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  init       Initialize envx for the current project.")
	fmt.Fprintln(w, "  set        Set or update an environment variable.")
	fmt.Fprintln(w, "  list       List environment variables.")
	fmt.Fprintln(w, "  run        Run a command with injected variables.")
	fmt.Fprintln(w, "  env        Manage named environments (create/use/delete/list).")
	fmt.Fprintln(w, "  help       Show this help.")
}
