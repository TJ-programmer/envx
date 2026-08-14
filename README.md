# envx

`envx` is a project-local environment variable manager: a faster, safer, drop-in replacement for `.env` files. It stores everything inside `<project>/.envx/`, encrypts secrets at rest with a project-local key, injects variables into subprocesses, and never touches your global shell state.

## Why envx

- **Project-specific, like `.env`** — all state lives in one `.envx/` directory; nothing global, no daemon, no network, no third-party.
- **Safer by default** — plain values by default (exact `.env` semantics); `--secret` encrypts values at rest with a project-local Fernet key.
- **No manual `.gitignore`** — secrets and keys live inside `.envx/`, so a single ignored directory is all you need.
- **Fast** — single Go binary, ~10ms startup.
- **No manual `.gitignore`** — `envx init` adds `.envx/` to your `.gitignore` automatically (opt out with `--no-gitignore`).
- **One command to migrate** — `envx init && envx import .env`.

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
envx init [--env dev] [--force] [--no-gitignore] [--root DIR]
envx set KEY VALUE [--env ENV] [--secret|--plain]
envx get KEY [--env ENV] [--show-secret]
envx list [--env ENV] [--show-secrets] [--format table|json]
envx unset KEY [--env ENV]
envx run [--env ENV] [--shell "cmd | pipe"] -- <command>...
envx env create|use|delete|list
envx import FILE.env [--env ENV]
envx export [--env ENV] [--format shell|dotenv|json]
envx diff ENV_A ENV_B
envx doctor
envx config get|set KEY [VALUE]
envx completions bash|zsh|fish|powershell
```

All commands accept `--root DIR` to point at a project instead of auto-discovering it by walking up to the nearest `.git/` or `.envx/` directory.

- `get` prints a value; secrets are redacted as `********` unless `--show-secret` is given (useful for scripts).
- `import` reads a `.env` file and stores each variable; values whose names look sensitive (e.g. `API_TOKEN`, `DB_PASSWORD`) are stored encrypted.
- `export` writes variables in `shell` (`export KEY='value'`), `dotenv` (`KEY=value`), or `json` form for piping to other tools.
- `diff` shows keys and values that differ between two environments.
- `doctor` checks project health: key present, config valid, secrets stored in plaintext that look sensitive, and whether `.envx/` is gitignored.
- `config` manages project-local settings such as `encryption.default_encrypt`; when enabled, plain `set` calls are stored encrypted automatically.
- `completions` prints shell completion scripts.

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
