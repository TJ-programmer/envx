//go:build windows

package run

import (
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
