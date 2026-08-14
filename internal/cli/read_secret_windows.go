//go:build windows

package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

var (
	kernel32            = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode  = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode  = kernel32.NewProc("SetConsoleMode")
	enableEchoInputMode = uint32(0x0004)
)

func readSecret(prompt string, in io.Reader, errOut io.Writer) (string, error) {
	fmt.Fprint(errOut, prompt)
	file, ok := in.(*os.File)
	if !ok {
		return readSecretLine(in)
	}
	defer fmt.Fprintln(errOut)

	handle := uintptr(file.Fd())
	var mode uint32
	r1, _, _ := procGetConsoleMode.Call(handle, uintptr(unsafe.Pointer(&mode)))
	if r1 == 0 {
		return readSecretLine(file)
	}
	procSetConsoleMode.Call(handle, uintptr(mode&^enableEchoInputMode))
	defer procSetConsoleMode.Call(handle, uintptr(mode))
	return readSecretLine(file)
}

func readSecretLine(in io.Reader) (string, error) {
	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}
