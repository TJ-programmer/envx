package bootstrap

import (
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
