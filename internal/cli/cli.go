package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

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
	case "get":
		return cmdGet(args[1:], stdout, stderr)
	case "list":
		return cmdList(args[1:], stdout, stderr)
	case "unset":
		return cmdUnset(args[1:], stdout, stderr)
	case "run":
		return cmdRun(args[1:], stdout, stderr)
	case "env":
		return cmdEnv(args[1:], stdout, stderr)
	case "import":
		return cmdImport(args[1:], stdout, stderr)
	case "export":
		return cmdExport(args[1:], stdout, stderr)
	case "diff":
		return cmdDiff(args[1:], stdout, stderr)
	case "doctor":
		return cmdDoctor(args[1:], stdout, stderr)
	case "config":
		return cmdConfig(args[1:], stdout, stderr)
	case "completions":
		return cmdCompletions(args[1:], stdout, stderr)
	case "help", "-h", "--help":
		printUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "Error: unknown command '%s'\n", args[0])
		printUsage(stderr)
		return 1
	}
}

func printError(stderr io.Writer, err error) int {
	fmt.Fprintf(stderr, "Error: %v\n", err)
	return 1
}

func splitRootFlag(args []string) (string, []string) {
	root := ""
	rest := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		if args[i] == "--root" {
			if i+1 < len(args) {
				root = args[i+1]
			}
			i++
			continue
		}
		rest = append(rest, args[i])
	}
	return root, rest
}

func projectService(root string) *core.EnvxService {
	if root == "" {
		root = bootstrap.DiscoverRoot()
	}
	return bootstrap.BuildService(root)
}

func cwdService(root string) *core.EnvxService {
	if root == "" {
		root, _ = os.Getwd()
	}
	return bootstrap.BuildService(root)
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

func renderDiffRows(rows []core.DiffRow, envA, envB string, stdout io.Writer) {
	if len(rows) == 0 {
		fmt.Fprintf(stdout, "No differences between '%s' and '%s'.\n", envA, envB)
		return
	}
	widths := []int{len("KEY"), len(envA), len(envB)}
	for _, row := range rows {
		widths[0] = maxInt(widths[0], len(row.Key))
		widths[1] = maxInt(widths[1], len(row.ValueA))
		widths[2] = maxInt(widths[2], len(row.ValueB))
	}
	fmt.Fprintf(stdout, "%-*s  %-*s  %-*s\n", widths[0], "KEY", widths[1], envA, widths[2], envB)
	for _, row := range rows {
		fmt.Fprintf(stdout, "%-*s  %-*s  %-*s\n", widths[0], row.Key, widths[1], row.ValueA, widths[2], row.ValueB)
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
	fmt.Fprintln(w, "  init         Initialize envx for the current project.")
	fmt.Fprintln(w, "  set          Set or update an environment variable.")
	fmt.Fprintln(w, "  get          Print the value of an environment variable.")
	fmt.Fprintln(w, "  list         List environment variables.")
	fmt.Fprintln(w, "  unset        Remove an environment variable.")
	fmt.Fprintln(w, "  run          Run a command with injected variables.")
	fmt.Fprintln(w, "  env          Manage named environments (create/use/delete/list).")
	fmt.Fprintln(w, "  import       Import variables from a .env file.")
	fmt.Fprintln(w, "  export       Export variables as shell, dotenv, or JSON.")
	fmt.Fprintln(w, "  diff         Compare two environments.")
	fmt.Fprintln(w, "  doctor       Diagnose project health (gitignore, secrets, key).")
	fmt.Fprintln(w, "  config       View or change project-local settings.")
	fmt.Fprintln(w, "  completions  Generate shell completions.")
	fmt.Fprintln(w, "  help         Show this help.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Global option:")
	fmt.Fprintln(w, "  --root DIR   Use DIR as the project root instead of auto-discovery.")
}
