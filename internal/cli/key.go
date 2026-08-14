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
	rest := args[1:]
	root, rest := splitRootFlag(rest)

	switch sub {
	case "status":
		if len(rest) != 0 {
			return printError(stderr, fmt.Errorf("unknown option '%s'", rest[0]))
		}
		paths := config.ResolveProjectPaths(resolveRoot(root))
		fmt.Fprintf(stdout, "Backend: file\n")
		fmt.Fprintf(stdout, "Key file: %s\n", paths.KeyPath)
		fmt.Fprintf(stdout, "Present: %t\n", fileExists(paths.KeyPath))
		if !fileExists(paths.KeyPath) && fileExists(paths.LegacyKeyPath) {
			fmt.Fprintf(stdout, "Legacy key present: true\n")
		}
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
		fmt.Fprintf(stdout, "Rotated encryption key; re-encrypted %d secret(s). Previous key saved to '.envx/key.old.bin'.\n", count)
		return 0
	case "export":
		if len(rest) != 0 {
			return printError(stderr, fmt.Errorf("unknown option '%s'", rest[0]))
		}
		paths := config.ResolveProjectPaths(resolveRoot(root))
		if !fileExists(paths.KeyPath) {
			return printError(stderr, errors.New("no key found; run 'envx init' first"))
		}
		data, err := os.ReadFile(paths.KeyPath)
		if err != nil {
			return printError(stderr, err)
		}
		fmt.Fprintln(stdout, strings.TrimSpace(string(data)))
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
		fmt.Fprintln(stdout, "Imported key to '.envx/key.bin' (previous key backed up to '.envx/key.old.bin').")
		return 0
	default:
		return printError(stderr, fmt.Errorf("unknown key subcommand '%s'", sub))
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
