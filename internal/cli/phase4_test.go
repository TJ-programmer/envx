package cli

import (
	"strings"
	"testing"
)

func runSilent(args ...string) (string, string, int) {
	var stdout, stderr strings.Builder
	code := Run(args, &stdout, &stderr)
	return stdout.String(), stderr.String(), code
}

func TestVersionFlag(t *testing.T) {
	for _, args := range [][]string{{"--version"}, {"-V"}, {"version"}} {
		out, _, code := runSilent(args...)
		if code != 0 {
			t.Fatalf("%v: exit code %d, want 0", args, code)
		}
		if !strings.Contains(out, "envx ") || !strings.Contains(out, "schema v") {
			t.Fatalf("%v: unexpected version output: %q", args, out)
		}
	}
}

func TestHelpCommand(t *testing.T) {
	usage, _, code := runSilent("help")
	if code != 0 || !strings.Contains(usage, "envx <command>") {
		t.Fatalf("help: code=%d output=%q", code, usage)
	}
	getUsage, _, code := runSilent("help", "get")
	if code != 0 || !strings.Contains(getUsage, "envx get KEY") {
		t.Fatalf("help get: code=%d output=%q", code, getUsage)
	}
	unknown, _, code := runSilent("help", "bogus")
	if code != 0 || !strings.Contains(unknown, "envx <command>") {
		t.Fatalf("help bogus: code=%d output=%q", code, unknown)
	}
}

func TestCommandHelpFlag(t *testing.T) {
	for _, cmd := range []string{"init", "set", "get", "list", "unset", "run", "shell", "copy", "env", "import", "export", "diff", "doctor", "config", "key", "web", "completions"} {
		out, _, code := runSilent(cmd, "--help")
		if code != 0 {
			t.Fatalf("%s --help: exit code %d, want 0", cmd, code)
		}
		if !strings.HasPrefix(out, "usage: envx ") {
			t.Fatalf("%s --help: expected usage line, got %q", cmd, out)
		}
	}
}
