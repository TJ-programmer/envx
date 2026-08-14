package keyring

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"envx/internal/config"
	"envx/internal/errs"
)

type LocalKeyProvider struct {
	paths config.ProjectPaths
}

func NewLocalKeyProvider(paths config.ProjectPaths) *LocalKeyProvider {
	return &LocalKeyProvider{paths: paths}
}

func (k *LocalKeyProvider) EnsureKey() error {
	if err := os.MkdirAll(k.paths.ConfigDir, 0o700); err != nil {
		return err
	}
	if k.exists(k.paths.KeyPath) {
		_, err := k.loadFrom(k.paths.KeyPath)
		return err
	}
	if k.exists(k.paths.LegacyKeyPath) {
		raw, err := os.ReadFile(k.paths.LegacyKeyPath)
		if err != nil {
			return err
		}
		if _, err := decodeKey(raw); err != nil {
			return err
		}
		if err := os.WriteFile(k.paths.KeyPath, raw, 0o600); err != nil {
			return err
		}
		return k.restrict(k.paths.KeyPath)
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return err
	}
	encoded := base64.URLEncoding.EncodeToString(raw)
	if err := os.WriteFile(k.paths.KeyPath, []byte(encoded), 0o600); err != nil {
		return err
	}
	return k.restrict(k.paths.KeyPath)
}

func (k *LocalKeyProvider) LoadKey() ([]byte, error) {
	if k.exists(k.paths.KeyPath) {
		return k.loadFrom(k.paths.KeyPath)
	}
	if k.exists(k.paths.LegacyKeyPath) {
		return k.loadFrom(k.paths.LegacyKeyPath)
	}
	return nil, fmt.Errorf("%w: encryption key not found. Run 'envx init' or restore '.envx/key.bin'.", errs.ErrSecretKey)
}

func (k *LocalKeyProvider) exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (k *LocalKeyProvider) loadFrom(path string) ([]byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return decodeKey(raw)
}

func decodeKey(raw []byte) ([]byte, error) {
	key, err := base64.URLEncoding.DecodeString(strings.TrimSpace(string(raw)))
	if err != nil || len(key) != 32 {
		return nil, fmt.Errorf("%w: encryption key is invalid or corrupted", errs.ErrSecretKey)
	}
	return key, nil
}

func (k *LocalKeyProvider) restrict(path string) error {
	if err := os.Chmod(path, 0o600); err != nil {
		return nil
	}
	return nil
}
