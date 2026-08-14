package clipboard

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Write copies content to the system clipboard without ever printing it to
// the terminal. It is a variable so tests can substitute a fake sink.
var Write = writeToClipboard

func writeToClipboard(content string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("clip.exe")
	default:
		cmd = clipboardCommand()
		if cmd == nil {
			return errors.New("no clipboard tool found (install wl-copy, xclip, xsel, or pbcopy)")
		}
	}
	cmd.Stdin = bytes.NewBufferString(content)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("clipboard copy failed: %v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func clipboardCommand() *exec.Cmd {
	for _, name := range []string{"wl-copy", "xclip", "xsel", "pbcopy"} {
		if path, err := exec.LookPath(name); err == nil {
			return exec.Command(path)
		}
	}
	return nil
}
