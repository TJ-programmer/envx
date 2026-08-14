# envx

`envx` is a project-local environment variable manager: a faster, safer, drop-in replacement for `.env` files. It stores everything inside `<project>/.envx/`, encrypts secrets at rest with a project-local key, injects variables into subprocesses, and never touches your global shell state.

## Why envx

- **Project-specific, like `.env`** — all state lives in one `.envx/` directory; nothing global, no daemon, no network, no third-party.
- **Safer by default** — plain values by default (exact `.env` semantics); `--secret` encrypts values at rest with a project-local Fernet key.
- **No manual `.gitignore`** — secrets and keys live inside `.envx/`, so a single ignored directory is all you need.
- **Fast** — single Go binary, ~10ms startup.
- **One command to migrate** — `envx init && envx import .env` (import lands in a later phase).

## Install

```bash
go install github.com/envx/envx/cmd/envx@latest
```

Or build from source:

```bash
go build -o bin/envx ./cmd/envx
```

## Quickstart

```bash
envx init
envx set PORT 8000
envx set API_KEY my-secret --secret
envx list
envx run -- python app.py
```

## Commands

```bash
envx init [--env dev] [--force]
envx set KEY VALUE [--env ENV] [--secret|--plain]
envx list [--env ENV] [--show-secrets] [--format table|json]
envx run [--env ENV] [--shell "cmd | pipe"] -- <command>...
envx env create|use|delete|list
```

## Storage

```
.envx/config.json          versioned config (schema v2)
.envx/config.backup.json   previous config before each write
.envx/config.lock          write lock
.envx/key.bin              encryption key (gitignored)
```

Encrypted values are stored as `enc:<fernet-token>` and redacted by default in `envx list`. Use `--show-secrets` to reveal them.

## Compatibility

The Go binary reads configs created by the earlier Python prototype (legacy schema and `key.key`), migrates them to the current schema in memory, and decrypts existing secrets using the same Fernet wire format. The Python reference implementation is kept in `legacy-python/`.

## Development

```bash
go test ./...
go vet ./...
gofmt -l .
```
