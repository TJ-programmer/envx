package cli

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"envx/internal/config"
	"envx/internal/keyring"
)

func cmdKey(args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 {
		return printError(stderr, errors.New("usage: envx key status|rotate|export|import"))
	}
	sub := args[0]
	if sub == "--help" || sub == "-h" {
		printCommandUsage(stdout, "key")
		return 0
	}
	rest := args[1:]
	root, rest := splitRootFlag(rest)
	if checkHelp(rest, stdout, "key") {
		return 0
	}

	switch sub {
	case "status":
		if len(rest) != 0 {
			return printError(stderr, fmt.Errorf("unknown option '%s'", rest[0]))
		}
		paths := config.ResolveProjectPaths(resolveRoot(root))
		provider := keyring.NewLocalKeyProvider(paths)
		backend := provider.Backend()
		fmt.Fprintf(stdout, "Backend: %s\n", backend)
		if backend == "keyring" {
			fmt.Fprintf(stdout, "Location: %s\n", provider.Location())
		} else {
			fmt.Fprintf(stdout, "Key file: %s\n", paths.KeyPath)
			if !provider.HasKey() && fileExists(paths.LegacyKeyPath) {
				fmt.Fprintf(stdout, "Legacy key present: true\n")
			}
		}
		fmt.Fprintf(stdout, "Present: %t\n", provider.HasKey())
		return 0
	case "rotate":
		if len(rest) != 0 {
			return printError(stderr, fmt.Errorf("unknown option '%s'", rest[0]))
		}
		service := projectService(root)
		count, err := service.RotateKey()
		if err != nil {
			return printError(stderr, err)
		}
		backend := keyring.NewLocalKeyProvider(config.ResolveProjectPaths(resolveRoot(root))).Backend()
		msg := "Rotated encryption key; re-encrypted %d secret(s)."
		if backend == "file" {
			msg += " Previous key saved to '.envx/key.old.bin'."
		} else {
			msg += " Previous key retained in the OS keyring."
		}
		fmt.Fprintf(stdout, msg+"\n", count)
		return 0
	case "export":
		if len(rest) != 0 {
			return printError(stderr, fmt.Errorf("unknown option '%s'", rest[0]))
		}
		paths := config.ResolveProjectPaths(resolveRoot(root))
		provider := keyring.NewLocalKeyProvider(paths)
		key, err := provider.LoadKey()
		if err != nil {
			return printError(stderr, err)
		}
		fmt.Fprintln(stdout, base64.URLEncoding.EncodeToString(key))
		return 0
	case "import":
		if len(rest) < 1 {
			return printError(stderr, errors.New("usage: envx key import FILE [--root DIR]"))
		}
		path := rest[0]
		if len(rest) > 1 {
			return printError(stderr, fmt.Errorf("unknown option '%s'", rest[1]))
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return printError(stderr, err)
		}
		key := strings.TrimSpace(string(raw))
		decoded, err := base64.URLEncoding.DecodeString(key)
		if err != nil || len(decoded) != 32 {
			return printError(stderr, errors.New("invalid key: expected a 32-byte URL-safe base64 Fernet key"))
		}
		paths := config.ResolveProjectPaths(resolveRoot(root))
		provider := keyring.NewLocalKeyProvider(paths)
		if err := provider.WriteKey(key); err != nil {
			return printError(stderr, err)
		}
		if provider.Backend() == "keyring" {
			fmt.Fprintln(stdout, "Imported key to the OS keyring (previous key retained as '...:old').")
		} else {
			fmt.Fprintln(stdout, "Imported key to '.envx/key.bin' (previous key backed up to '.envx/key.old.bin').")
		}
		return 0
	default:
		return printError(stderr, fmt.Errorf("unknown key subcommand '%s'", sub))
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
