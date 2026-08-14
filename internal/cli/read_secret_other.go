//go:build !windows

package cli

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func readSecret(prompt string, in io.Reader, errOut io.Writer) (string, error) {
	fmt.Fprint(errOut, prompt)
	defer fmt.Fprintln(errOut)
	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}
