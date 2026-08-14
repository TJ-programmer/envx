//go:build windows

package run

import (
	"os"
	"os/exec"
	"syscall"
)

func buildShellCmd(shellCmd string) *exec.Cmd {
	path, err := exec.LookPath("cmd")
	if err != nil {
		path = "cmd"
	}
	return &exec.Cmd{
		Path: path,
		SysProcAttr: &syscall.SysProcAttr{
			CmdLine: `cmd /S /C "` + shellCmd + `"`,
		},
	}
}

func interactiveShell() (name string, args []string, extraEnv []string, cleanup func()) {
	if p, err := exec.LookPath("pwsh.exe"); err == nil {
		return p, []string{"-NoExit", "-Command", `function global:prompt { "(envx) ❯ " }`}, nil, nil
	}
	if p, err := exec.LookPath("powershell.exe"); err == nil {
		return p, []string{"-NoExit", "-Command", `function global:prompt { "(envx) ❯ " }`}, nil, nil
	}
	comspec := os.Getenv("ComSpec")
	if comspec == "" {
		comspec = "cmd.exe"
	}
	return comspec, nil, []string{"PROMPT=(envx) $G"}, nil
}
