package bootstrap

import (
	"os"
	"path/filepath"
	"strings"

	"envx/internal/config"
	"envx/internal/core"
	"envx/internal/crypto"
	"envx/internal/keyring"
	"envx/internal/run"
	"envx/internal/store"
)

func BuildService(baseDir string) *core.EnvxService {
	paths := config.ResolveProjectPaths(baseDir)
	return core.NewService(
		store.New(paths),
		crypto.NewCryptoManager(keyring.NewLocalKeyProvider(paths)),
		run.NewRunner(),
	)
}

func DiscoverRoot() string {
	start, err := os.Getwd()
	if err != nil {
		return start
	}
	dir, err := filepath.Abs(start)
	if err != nil {
		return start
	}
	home, _ := os.UserHomeDir()
	homeAbs, _ := filepath.Abs(home)
	for {
		if hasProjectMarker(dir) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return start
		}
		if homeAbs != "" && strings.EqualFold(filepath.Clean(dir), filepath.Clean(homeAbs)) {
			return start
		}
		dir = parent
	}
}

func hasProjectMarker(dir string) bool {
	for _, name := range []string{".envx", ".git"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			return true
		}
	}
	return false
}
