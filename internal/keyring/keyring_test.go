//go:build windows

package keyring

import (
	"encoding/base64"
	"os"
	"testing"

	"envx/internal/config"
)

const testConfigJSON = `{"version":2,"active_env":"dev","encryption":{"default_encrypt":false,"key_backend":"%s"},"migration":{"overlay_dotenv":false},"environments":{}}`

func newTestProvider(t *testing.T, backend string) *LocalKeyProvider {
	t.Helper()
	root := t.TempDir()
	paths := config.ResolveProjectPaths(root)
	writeConfig(t, paths, backend)
	provider := NewLocalKeyProvider(paths)
	t.Cleanup(func() {
		_ = credDelete(provider.credTarget())
		_ = credDelete(provider.credOldTarget())
	})
	return provider
}

func writeConfig(t *testing.T, paths config.ProjectPaths, backend string) {
	t.Helper()
	if err := os.MkdirAll(paths.ConfigDir, 0o700); err != nil {
		t.Fatal(err)
	}
	content := "{\"encryption\":{\"key_backend\":\"" + backend + "\"}}"
	if err := os.WriteFile(paths.ConfigPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureKeyWithKeyringBackend(t *testing.T) {
	provider := newTestProvider(t, "keyring")
	if err := provider.EnsureKey(); err != nil {
		t.Fatal(err)
	}
	if !provider.HasKey() {
		t.Fatal("expected key in OS keyring")
	}
	if provider.exists(provider.paths.KeyPath) {
		t.Fatal("key.bin should not exist with the keyring backend")
	}
	raw, err := provider.LoadKey()
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) != 32 {
		t.Fatalf("key length = %d, want 32", len(raw))
	}
}

func TestWriteKeyWithKeyringBackend(t *testing.T) {
	provider := newTestProvider(t, "keyring")
	if err := provider.EnsureKey(); err != nil {
		t.Fatal(err)
	}
	newRaw := make([]byte, 32)
	for i := range newRaw {
		newRaw[i] = byte(i)
	}
	newEncoded := base64.URLEncoding.EncodeToString(newRaw)
	if err := provider.WriteKey(newEncoded); err != nil {
		t.Fatal(err)
	}
	got, err := provider.LoadKey()
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(newRaw) {
		t.Fatal("key did not rotate in the keyring")
	}
	old, err := credRead(provider.credOldTarget())
	if err != nil {
		t.Fatal(err)
	}
	if string(old) == string(newRaw) {
		t.Fatal("old key backup should differ from the new key")
	}
	if len(old) != 32 {
		t.Fatalf("old key backup length = %d", len(old))
	}
}

func TestMigrateFileToKeyring(t *testing.T) {
	root := t.TempDir()
	paths := config.ResolveProjectPaths(root)
	if err := os.MkdirAll(paths.ConfigDir, 0o700); err != nil {
		t.Fatal(err)
	}
	raw := make([]byte, 32)
	raw[0] = 0xAA
	if err := os.WriteFile(paths.KeyPath, []byte(base64.URLEncoding.EncodeToString(raw)), 0o600); err != nil {
		t.Fatal(err)
	}

	provider := NewLocalKeyProvider(paths)
	writeConfig(t, paths, "keyring")
	t.Cleanup(func() {
		_ = credDelete(provider.credTarget())
		_ = credDelete(provider.credOldTarget())
	})

	if err := provider.EnsureKey(); err != nil {
		t.Fatal(err)
	}
	got, err := provider.LoadKey()
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(raw) {
		t.Fatal("key changed during file->keyring migration")
	}
	if provider.exists(paths.KeyPath) {
		t.Fatal("key.bin should be removed after migrating to the keyring")
	}
}

func TestMigrateKeyringToFile(t *testing.T) {
	provider := newTestProvider(t, "keyring")
	if err := provider.EnsureKey(); err != nil {
		t.Fatal(err)
	}

	writeConfig(t, provider.paths, "file")
	raw, err := provider.LoadKey()
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) != 32 {
		t.Fatalf("key length = %d, want 32", len(raw))
	}
	if !provider.exists(provider.paths.KeyPath) {
		t.Fatal("key.bin should exist after migrating back to the file backend")
	}
	if _, err := credRead(provider.credTarget()); !credNotFound(err) {
		t.Fatal("keyring entry should be removed after migrating back to file")
	}
}
