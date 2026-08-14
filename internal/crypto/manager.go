package crypto

import (
	"encoding/base64"
	"fmt"
	"strings"

	"envx/internal/errs"
)

type KeyProvider interface {
	EnsureKey() error
	LoadKey() ([]byte, error)
}

type CryptoManager struct {
	provider KeyProvider
}

func NewCryptoManager(provider KeyProvider) *CryptoManager {
	return &CryptoManager{provider: provider}
}

func (c *CryptoManager) EnsureKey() error {
	return c.provider.EnsureKey()
}

func (c *CryptoManager) Encrypt(value string) (string, error) {
	key, err := c.provider.LoadKey()
	if err != nil {
		return "", err
	}
	fernet, err := NewFernet(key)
	if err != nil {
		return "", err
	}
	token, err := fernet.Encrypt([]byte(value))
	if err != nil {
		return "", err
	}
	return EncryptedPrefix + base64.URLEncoding.EncodeToString(token), nil
}

func (c *CryptoManager) Decrypt(value string) (string, error) {
	if !strings.HasPrefix(value, EncryptedPrefix) {
		return value, nil
	}
	key, err := c.provider.LoadKey()
	if err != nil {
		return "", err
	}
	fernet, err := NewFernet(key)
	if err != nil {
		return "", err
	}
	token, err := base64.URLEncoding.DecodeString(value[len(EncryptedPrefix):])
	if err != nil {
		return "", fmt.Errorf("%w: failed to decrypt a stored secret. The key may be missing or mismatched.", errs.ErrSecretKey)
	}
	plain, err := fernet.Decrypt(token)
	if err != nil {
		return "", fmt.Errorf("%w: failed to decrypt a stored secret. The key may be missing or mismatched.", errs.ErrSecretKey)
	}
	return string(plain), nil
}
