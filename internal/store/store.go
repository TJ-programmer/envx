package store

import (
	"encoding/json"
	"fmt"
	"os"

	"envx/internal/config"
	"envx/internal/core"
	"envx/internal/errs"
)

type Store struct {
	paths config.ProjectPaths
}

func New(paths config.ProjectPaths) *Store {
	return &Store{paths: paths}
}

func (s *Store) Exists() bool {
	_, err := os.Stat(s.paths.ConfigPath)
	return err == nil
}

func (s *Store) Initialize(cfg *core.ConfigFile, force bool) error {
	if s.Exists() && !force {
		return errs.ErrAlreadyInitialized
	}
	return s.Save(cfg)
}

func (s *Store) Load() (*core.ConfigFile, error) {
	if !s.Exists() {
		return nil, errs.ErrProjectNotInitialized
	}

	raw, err := loadJSON(s.paths.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("%w: config file contains invalid JSON. Restore '.envx/config.backup.json' or re-run 'envx init --force'.", errs.ErrConfigCorrupted)
	}

	normalized, err := migrateConfig(raw)
	if err != nil {
		return nil, err
	}

	cfg := &core.ConfigFile{}
	data, err := json.Marshal(normalized)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("%w: config file is corrupted. Restore '.envx/config.backup.json' or re-run 'envx init --force'.", errs.ErrConfigCorrupted)
	}

	return cfg, core.ValidateConfig(cfg)
}

func (s *Store) Save(cfg *core.ConfigFile) error {
	if err := core.ValidateConfig(cfg); err != nil {
		return err
	}
	if err := os.MkdirAll(s.paths.ConfigDir, 0o755); err != nil {
		return err
	}

	lock := NewFileLock(s.paths.LockPath)
	if err := lock.Lock(); err != nil {
		return err
	}
	defer lock.Unlock()

	if s.Exists() {
		data, err := os.ReadFile(s.paths.ConfigPath)
		if err != nil {
			return err
		}
		if err := os.WriteFile(s.paths.BackupPath, data, 0o644); err != nil {
			return err
		}
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	tmpPath := s.paths.ConfigPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmpPath, s.paths.ConfigPath)
}

func loadJSON(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}
