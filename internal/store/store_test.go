package store_test

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"envx/internal/config"
	"envx/internal/core"
	"envx/internal/crypto"
	"envx/internal/errs"
	"envx/internal/keyring"
	"envx/internal/store"
)

func buildStore(dir string) (config.ProjectPaths, *store.Store, *crypto.CryptoManager) {
	paths := config.ResolveProjectPaths(dir)
	return paths, store.New(paths), crypto.NewCryptoManager(keyring.NewLocalKeyProvider(paths))
}

func TestConfigRoundTrip(t *testing.T) {
	paths, s, _ := buildStore(t.TempDir())
	cfg := core.NewConfigFile("dev")
	if err := s.Save(cfg); err != nil {
		t.Fatal(err)
	}

	loaded, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Version != core.SchemaVersion {
		t.Errorf("version = %d, want %d", loaded.Version, core.SchemaVersion)
	}
	if loaded.ActiveEnv != "dev" {
		t.Errorf("active_env = %q", loaded.ActiveEnv)
	}
	if _, ok := loaded.Environments["dev"]; !ok {
		t.Error("dev environment missing")
	}
	if _, err := os.Stat(paths.ConfigPath); err != nil {
		t.Errorf("config file missing: %v", err)
	}
}

func TestBackupIsWrittenOnSecondSave(t *testing.T) {
	paths, s, _ := buildStore(t.TempDir())
	cfg := core.NewConfigFile("dev")
	if err := s.Save(cfg); err != nil {
		t.Fatal(err)
	}
	cfg.Metadata["updated"] = true
	if err := s.Save(cfg); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(paths.BackupPath); err != nil {
		t.Fatalf("backup not written: %v", err)
	}
	backup, err := os.ReadFile(paths.BackupPath)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(backup, &raw); err != nil {
		t.Fatal(err)
	}
	if raw["active_env"] != "dev" {
		t.Errorf("backup active_env = %v", raw["active_env"])
	}
}

func TestLegacyConfigIsMigrated(t *testing.T) {
	paths, s, _ := buildStore(t.TempDir())
	if err := os.MkdirAll(paths.ConfigDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(paths.ConfigPath, []byte(`{"active_env":"dev","environments":{"dev":{"PORT":"8000"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Version != core.SchemaVersion {
		t.Errorf("version = %d, want %d", cfg.Version, core.SchemaVersion)
	}
	port, ok := cfg.Environments["dev"].Variables["PORT"]
	if !ok {
		t.Fatal("PORT variable missing after migration")
	}
	if port.Value != "8000" {
		t.Errorf("PORT = %q", port.Value)
	}
}

func TestCorruptConfigRaisesActionableError(t *testing.T) {
	paths, s, _ := buildStore(t.TempDir())
	if err := os.MkdirAll(paths.ConfigDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(paths.ConfigPath, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := s.Load()
	if !errors.Is(err, errs.ErrConfigCorrupted) {
		t.Fatalf("err = %v", err)
	}
	if !strings.Contains(err.Error(), "invalid JSON") {
		t.Fatalf("error should mention invalid JSON: %v", err)
	}
}

func TestCryptoInvalidKeyFailsClosed(t *testing.T) {
	dir := t.TempDir()
	paths, _, cryptoManager := buildStore(dir)
	if err := os.MkdirAll(paths.ConfigDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(paths.KeyPath, []byte("broken"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := cryptoManager.Decrypt("enc:abc"); !errors.Is(err, errs.ErrSecretKey) {
		t.Fatalf("err = %v", err)
	}
}
