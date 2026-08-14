package gitignore

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	marker = "# envx"
	entry  = ".envx/"
)

func Ensure(dir string, skip bool) error {
	if skip {
		return nil
	}
	gitignorePath := filepath.Join(dir, ".gitignore")
	_, gitErr := os.Stat(filepath.Join(dir, ".git"))
	_, ignoreErr := os.Stat(gitignorePath)
	if os.IsNotExist(gitErr) && os.IsNotExist(ignoreErr) {
		return nil
	}

	content := ""
	if data, err := os.ReadFile(gitignorePath); err == nil {
		content = string(data)
	}
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == entry {
			return nil
		}
	}
	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	content += "\n" + marker + "\n" + entry + "\n"
	return os.WriteFile(gitignorePath, []byte(content), 0o644)
}

func IsManaged(dir string) (bool, error) {
	gitignorePath := filepath.Join(dir, ".gitignore")
	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == entry {
			return true, nil
		}
	}
	return false, nil
}
