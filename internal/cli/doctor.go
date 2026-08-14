package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"envx/internal/bootstrap"
	"envx/internal/config"
	"envx/internal/core"
	"envx/internal/gitignore"
	"envx/internal/keyring"
	"envx/internal/store"
)

func cmdDoctor(args []string, stdout, stderr io.Writer) int {
	root, rest := splitRootFlag(args)
	if checkHelp(rest, stdout, "doctor") {
		return 0
	}
	if len(rest) > 0 {
		return printError(stderr, fmt.Errorf("unknown option '%s'", rest[0]))
	}
	if root == "" {
		root = bootstrap.DiscoverRoot()
	}
	paths := config.ResolveProjectPaths(root)

	storeCfg := store.New(paths)
	if !storeCfg.Exists() {
		fmt.Fprintf(stdout, "[fail] Project is not initialized. Run 'envx init'.\n")
		return 1
	}
	fmt.Fprintf(stdout, "[ok]   Project initialized at %s\n", paths.Root)

	cfg, err := storeCfg.Load()
	if err != nil {
		fmt.Fprintf(stdout, "[fail] Config invalid: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "[ok]   Config valid (schema v%d, active env '%s')\n", cfg.Version, cfg.ActiveEnv)

	if cfg.Encryption.KeyBackend == "keyring" {
		provider := keyring.NewLocalKeyProvider(paths)
		if !keyring.Supported() {
			fmt.Fprintf(stdout, "[fail] Config selects the OS keyring backend, which is not supported on this platform\n")
			return 1
		}
		if provider.HasKey() {
			fmt.Fprintf(stdout, "[ok]   Encryption key present in the OS keyring\n")
		} else {
			fmt.Fprintf(stdout, "[fail] Encryption key missing from the OS keyring\n")
		}
		if _, err := os.Stat(paths.KeyPath); err == nil {
			fmt.Fprintf(stdout, "[warn] '.envx/key.bin' still exists; delete it once you confirm the keyring entry works\n")
		}
	} else {
		if _, err := os.Stat(paths.KeyPath); err == nil {
			fmt.Fprintf(stdout, "[ok]   Encryption key present\n")
		} else if _, err := os.Stat(paths.LegacyKeyPath); err == nil {
			fmt.Fprintf(stdout, "[warn] Using legacy key file (key.key); run any write command to migrate\n")
		} else {
			fmt.Fprintf(stdout, "[fail] Encryption key missing\n")
		}
	}

	warnings := false
	for envName, environment := range cfg.Environments {
		for name, entry := range environment.Variables {
			if !entry.IsSecret && core.IsSensitiveName(name) {
				fmt.Fprintf(stdout, "[warn] '%s' in environment '%s' looks sensitive but is stored in plaintext (use 'envx set %s ... --secret')\n", name, envName, name)
				warnings = true
			}
		}
	}
	if !warnings {
		fmt.Fprintf(stdout, "[ok]   No sensitive-looking variables stored in plaintext\n")
	}

	gitDir := filepath.Join(paths.Root, ".git")
	if _, gitErr := os.Stat(gitDir); gitErr == nil {
		managed, checkErr := gitignore.IsManaged(paths.Root)
		if checkErr != nil {
			fmt.Fprintf(stdout, "[warn] Could not read .gitignore: %v\n", checkErr)
		} else if managed {
			fmt.Fprintf(stdout, "[ok]   .gitignore ignores .envx/\n")
		} else {
			fmt.Fprintf(stdout, "[warn] .gitignore does not ignore .envx/ (run 'envx init' to add it)\n")
			warnings = true
		}
	} else {
		fmt.Fprintf(stdout, "[ok]   Not a git repository, nothing to ignore\n")
	}

	if _, backupErr := os.Stat(paths.BackupPath); backupErr == nil {
		fmt.Fprintf(stdout, "[ok]   Backup present\n")
	} else {
		fmt.Fprintf(stdout, "[warn] No backup file yet (created after the first change)\n")
	}

	if warnings {
		fmt.Fprintf(stdout, "Doctor found warnings; review the items above.\n")
		return 1
	}
	fmt.Fprintf(stdout, "All checks passed.\n")
	return 0
}
