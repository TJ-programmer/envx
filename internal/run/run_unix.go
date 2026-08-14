//go:build !windows

package run

import "os/exec"

func buildShellCmd(shellCmd string) *exec.Cmd {
	return exec.Command("sh", "-c", shellCmd)
}
