package keyring

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"envx/internal/config"
	"envx/internal/errs"
)

const backendFile = "file"
const backendKeyring = "keyring"

type LocalKeyProvider struct {
	paths config.ProjectPaths
}

func NewLocalKeyProvider(paths config.ProjectPaths) *LocalKeyProvider {
	return &LocalKeyProvider{paths: paths}
}

func Supported() bool {
	return supportsKeyring()
}

func (k *LocalKeyProvider) Backend() string {
	data, err := os.ReadFile(k.paths.ConfigPath)
	if err != nil {
		return backendFile
	}
	var cfg struct {
		Encryption struct {
			KeyBackend string `json:"key_backend"`
		} `json:"encryption"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return backendFile
	}
	if cfg.Encryption.KeyBackend == backendKeyring {
		return backendKeyring
	}
	return backendFile
}

func (k *LocalKeyProvider) Location() string {
	return k.credTarget()
}

func (k *LocalKeyProvider) HasKey() bool {
	if k.Backend() == backendKeyring {
		_, err := k.loadCredKey()
		return err == nil
	}
	return k.exists(k.paths.KeyPath) || k.exists(k.paths.LegacyKeyPath)
}

func (k *LocalKeyProvider) EnsureKey() error {
	if err := os.MkdirAll(k.paths.ConfigDir, 0o700); err != nil {
		return err
	}

	if k.Backend() == backendKeyring {
		if _, err := k.loadCredKey(); err == nil {
			return nil
		} else if !credNotFound(err) {
			return err
		}
		if raw, err := k.loadFileKey(); err == nil {
			if err := k.saveCredKey(raw); err != nil {
				return err
			}
			return k.removeFileKey()
		}
		return k.saveCredKey(k.newKey())
	}

	if !k.exists(k.paths.KeyPath) && !k.exists(k.paths.LegacyKeyPath) {
		if raw, err := k.loadCredKey(); err == nil {
			if err := k.saveFileKey(raw); err != nil {
				return err
			}
			return credDelete(k.credTarget())
		} else if !credNotFound(err) {
			return err
		}
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

	return k.saveFileKey(k.newKey())
}

func (k *LocalKeyProvider) LoadKey() ([]byte, error) {
	if k.Backend() == backendKeyring {
		if raw, err := k.loadCredKey(); err == nil {
			return raw, nil
		} else if !credNotFound(err) {
			return nil, err
		}
		if raw, err := k.loadFileKey(); err == nil {
			if err := k.saveCredKey(raw); err != nil {
				return nil, err
			}
			if err := k.removeFileKey(); err != nil {
				return nil, err
			}
			return raw, nil
		}
		return nil, fmt.Errorf("%w: encryption key not found. Run 'envx init' or import the key with 'envx key import'.", errs.ErrSecretKey)
	}

	if !k.exists(k.paths.KeyPath) && !k.exists(k.paths.LegacyKeyPath) {
		if raw, err := k.loadCredKey(); err == nil {
			if err := k.saveFileKey(raw); err != nil {
				return nil, err
			}
			return raw, credDelete(k.credTarget())
		}
	}

	if k.exists(k.paths.KeyPath) {
		return k.loadFrom(k.paths.KeyPath)
	}
	if k.exists(k.paths.LegacyKeyPath) {
		return k.loadFrom(k.paths.LegacyKeyPath)
	}
	return nil, fmt.Errorf("%w: encryption key not found. Run 'envx init' or restore '.envx/key.bin'.", errs.ErrSecretKey)
}

func (k *LocalKeyProvider) WriteKey(raw string) error {
	decoded, err := decodeKey([]byte(raw))
	if err != nil {
		return err
	}

	if k.Backend() == backendKeyring {
		if oldKey, err := k.loadCredKey(); err == nil {
			if err := credDelete(k.credOldTarget()); err != nil && !credNotFound(err) {
				return err
			}
			if err := credWrite(k.credOldTarget(), oldKey); err != nil {
				return err
			}
		} else if !credNotFound(err) {
			return err
		}
		if err := k.saveCredKey(decoded); err != nil {
			return err
		}
		return k.removeFileKey()
	}

	if k.exists(k.paths.KeyPath) {
		oldData, err := os.ReadFile(k.paths.KeyPath)
		if err != nil {
			return err
		}
		if err := os.WriteFile(k.paths.OldKeyPath, oldData, 0o600); err != nil {
			return err
		}
		if err := k.restrict(k.paths.OldKeyPath); err != nil {
			return err
		}
	}
	return k.saveFileKey(decoded)
}

func (k *LocalKeyProvider) newKey() []byte {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		panic(err)
	}
	return raw
}

func (k *LocalKeyProvider) credTarget() string {
	return "envx:" + k.paths.Root
}

func (k *LocalKeyProvider) credOldTarget() string {
	return "envx:" + k.paths.Root + ":old"
}

func (k *LocalKeyProvider) loadFileKey() ([]byte, error) {
	if k.exists(k.paths.KeyPath) {
		return k.loadFrom(k.paths.KeyPath)
	}
	if k.exists(k.paths.LegacyKeyPath) {
		return k.loadFrom(k.paths.LegacyKeyPath)
	}
	return nil, fmt.Errorf("%w: no file key found", errs.ErrSecretKey)
}

func (k *LocalKeyProvider) saveFileKey(raw []byte) error {
	encoded := base64.URLEncoding.EncodeToString(raw)
	if err := os.MkdirAll(k.paths.ConfigDir, 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(k.paths.KeyPath, []byte(encoded), 0o600); err != nil {
		return err
	}
	return k.restrict(k.paths.KeyPath)
}

func (k *LocalKeyProvider) removeFileKey() error {
	var firstErr error
	for _, path := range []string{k.paths.KeyPath, k.paths.LegacyKeyPath} {
		if k.exists(path) {
			if err := os.Remove(path); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

func (k *LocalKeyProvider) loadCredKey() ([]byte, error) {
	if !supportsKeyring() {
		return nil, errors.New(errKeyringUnsupported)
	}
	return credRead(k.credTarget())
}

func (k *LocalKeyProvider) saveCredKey(raw []byte) error {
	if !supportsKeyring() {
		return errors.New(errKeyringUnsupported)
	}
	return credWrite(k.credTarget(), raw)
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
