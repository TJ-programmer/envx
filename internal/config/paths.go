package config

import (
	"path/filepath"
)

const LegacyKeyFilename = "key.key"

type ProjectPaths struct {
	Root          string
	ConfigDir     string
	ConfigPath    string
	BackupPath    string
	LockPath      string
	KeyPath       string
	LegacyKeyPath string
}

func ResolveProjectPaths(baseDir string) ProjectPaths {
	root, _ := filepath.Abs(baseDir)
	configDir := filepath.Join(root, ".envx")
	return ProjectPaths{
		Root:          root,
		ConfigDir:     configDir,
		ConfigPath:    filepath.Join(configDir, "config.json"),
		BackupPath:    filepath.Join(configDir, "config.backup.json"),
		LockPath:      filepath.Join(configDir, "config.lock"),
		KeyPath:       filepath.Join(configDir, "key.bin"),
		LegacyKeyPath: filepath.Join(configDir, LegacyKeyFilename),
	}
}
