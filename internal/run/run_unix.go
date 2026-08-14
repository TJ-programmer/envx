//go:build !windows

package run

import (
	"os"
	"os/exec"
	"path/filepath"
)

func buildShellCmd(shellCmd string) *exec.Cmd {
	return exec.Command("sh", "-c", shellCmd)
}

func interactiveShell() (name string, args []string, extraEnv []string, cleanup func()) {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}
	base := filepath.Base(shell)
	switch base {
	case "bash":
		if f, err := os.CreateTemp("", "envx_bashrc_*.sh"); err == nil {
			content := "if [ -f \"$HOME/.bashrc\" ]; then . \"$HOME/.bashrc\"; fi\nPS1=\"\\n(envx) ❯ \"\n"
			if _, werr := f.WriteString(content); werr == nil {
				_ = f.Close()
				return shell, []string{"--rcfile", f.Name(), "-i"}, nil, func() { _ = os.Remove(f.Name()) }
			}
			_ = f.Close()
		}
		return shell, []string{"-i"}, nil, nil
	case "zsh":
		if tmp, err := os.MkdirTemp("", "envx_zsh_"); err == nil {
			rc := filepath.Join(tmp, ".zshrc")
			content := "if [ -f \"$HOME/.zshrc\" ]; then source \"$HOME/.zshrc\"; fi\nPROMPT=\"(envx) ❯ \"\n"
			if werr := os.WriteFile(rc, []byte(content), 0o600); werr == nil {
				return shell, []string{"-i"}, []string{"ZDOTDIR=" + tmp}, func() { _ = os.RemoveAll(tmp) }
			}
			_ = os.RemoveAll(tmp)
		}
		return shell, []string{"-i"}, nil, nil
	case "fish":
		return shell, []string{"-i", "-C", "function fish_prompt; echo -n '(envx) ❯ '; end"}, nil, nil
	default:
		return shell, []string{"-i"}, []string{"PS1=(envx) ❯ "}, nil
	}
}
