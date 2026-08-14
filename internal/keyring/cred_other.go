//go:build !windows

package keyring

import "errors"

const errKeyringUnsupported = "OS keyring backend is not supported on this platform"

var errCredentialNotFound = errors.New("credential not found")

func supportsKeyring() bool {
	return false
}

func credWrite(target string, blob []byte) error {
	return errors.New(errKeyringUnsupported)
}

func credRead(target string) ([]byte, error) {
	return nil, errors.New(errKeyringUnsupported)
}

func credDelete(target string) error {
	return errors.New(errKeyringUnsupported)
}

func credNotFound(err error) bool {
	return errors.Is(err, errCredentialNotFound)
}
