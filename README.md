<div align="center">

# ⚡ envx

**A faster, safer drop-in replacement for `.env` files.**

Project-local environment variables. Secrets encrypted at rest. Zero daemons, zero network, zero third-party.

![License: MIT](https://img.shields.io/badge/license-MIT-blue)
![Go](https://img.shields.io/badge/Go-1.26+-00ADD8)
![Platforms](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-6f42c1)
![Single static binary](https://img.shields.io/badge/single%20static%20binary-4cc61e)

```bash
curl -fsSL https://raw.githubusercontent.com/TJ-programmer/envx/main/scripts/install.sh | sh
```

</div>

---

## 🧭 What is envx?

`envx` keeps every environment variable for a project in one place — `<project>/.envx/` — and injects them into your commands, your shell, and your scripts.

It is **not** a daemon, a cloud service, or a framework. It is one small static binary that reads a folder.

- 🔒 **Plain by default, encrypted when it matters** — values are stored exactly like `.env` unless you mark them `--secret`, in which case they're encrypted at rest with a project-local Fernet key.
- 🌱 **Project-local** — all state lives in `.envx/`. Nothing global, nothing in your shell profile.
- 🤐 **Secrets stay out of git** — `envx init` adds `.envx/` to your `.gitignore` automatically.
- ⚡ **Fast** — single Go binary, ~10ms startup.
- 📦 **No dependencies, no framework** — prebuilt static binaries for Linux, macOS, and Windows (amd64 + arm64). Works with any language, any runtime, any framework.

---

## 🚀 Quickstart

```bash
envx init                              # create .envx/ + config
envx set PORT 8000                     # plain value
envx set API_KEY my-secret --secret    # encrypted value
envx list                              # table view (secrets redacted)
envx run -- python app.py              # run with variables injected
```

Under the hood that's `ENVX_ACTIVE=1 PORT=8000 API_KEY=... python app.py`.

```bash
envx shell                             # interactive shell with vars loaded
envx copy API_KEY                      # copy a secret to the clipboard
envx run --watch -- python app.py      # auto-restart when env changes
```

---

## 📦 Install

envx ships as a **single static binary** — no runtime, interpreter, or package manager required.

### Package managers

Install with whatever you already use — npm/pnpm/bun, Homebrew, or Scoop all work out of the box:

```bash
# npm / pnpm / bun (downloads the prebuilt binary for your platform)
npm install -g envx
pnpm add -g envx
bun add -g envx

# macOS / Linux
brew tap TJ-programmer/envx
brew install envx

# Windows
scoop bucket add envx https://github.com/TJ-programmer/scoop-envx
scoop install envx
```

### One-liners

```bash
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/TJ-programmer/envx/main/scripts/install.sh | sh

# Windows (PowerShell)
powershell -NoProfile -ExecutionPolicy Bypass -Command "iwr -useb https://raw.githubusercontent.com/TJ-programmer/envx/main/scripts/install.ps1 | iex"
```

### Go developers

```bash
go install github.com/TJ-programmer/envx/cmd/envx@latest
```

### Build from source

```bash
git clone https://github.com/TJ-programmer/envx
cd envx
go build -o bin/envx ./cmd/envx
```

> **Forks:** the installers and the npm package default to `TJ-programmer/envx`. Point them anywhere with `ENVX_REPO=owner/repo` (or `$env:ENVX_REPO` on Windows). The npm package also honors `ENVX_VERSION` to install a specific version.

### How releases work

Pushing a tag like `v0.5.0` triggers the [release workflow](.github/workflows/release.yml) (GitHub Actions + GoReleaser), which cross-compiles static binaries for every platform, attaches them plus a `checksums.txt` to a GitHub Release, and embeds the tag version via `-ldflags`. The installers, the npm package, and the Scoop manifest fetch the archive matching your OS/CPU and verify its SHA-256 checksum. After each release, publish the updated package to npm and refresh the Homebrew formula / Scoop hashes (see [CONTRIBUTING.md](CONTRIBUTING.md)).

---

## 🧠 Commands

```bash
envx init [--env dev] [--backend file|keyring] [--force] [--no-gitignore] [--root DIR]
envx set KEY [VALUE] [--env ENV] [--secret|--plain]
envx get KEY [--env ENV] [--show-secret]
envx list [--env ENV] [--show-secrets] [--format table|json]
envx unset KEY [--env ENV]
envx run [--env ENV] [--shell "cmd | pipe"] [--overlay] [--watch] -- <command>...
envx shell [--env ENV] [--shell CMD] [--overlay]
envx copy KEY [KEY...] [--env ENV]
envx env create|use|delete|list
envx import FILE.env [--env ENV]
envx export [--env ENV] [--format shell|dotenv|json]
envx diff ENV_A ENV_B
envx doctor
envx config get|set KEY [VALUE]
envx key status|rotate|export|import
envx web [--port PORT] [--no-open]
envx version
envx help [command]      # or: envx <command> --help
envx completions bash|zsh|fish|powershell
```

Every command accepts `--root DIR` to target a specific project instead of auto-discovering it by walking up to the nearest `.git/` or `.envx/` directory.

| Command | What it does |
|---|---|
| `get` | Prints a value; secrets redacted as `********` unless `--show-secret` (handy for scripts). |
| `import` | Reads a `.env` file; values with sensitive-looking names (`API_TOKEN`, `DB_PASSWORD`, …) are stored encrypted automatically. |
| `export` | Writes `shell` (`export KEY='value'`), `dotenv` (`KEY=value`), or `json` for piping to other tools. |
| `diff` | Shows keys and values that differ between two environments. |
| `doctor` | Health check: key present, config valid, suspicious plaintext secrets, `.envx/` gitignored. |
| `config` | Project settings like `encryption.default_encrypt`, `encryption.key_backend`, `migration.overlay_dotenv`. |
| `run --watch` | Restarts the child command whenever `.env`, `.envx/config.json`, or the key file changes — great for dev servers. |
| `shell` | Interactive subshell with variables loaded (`pwsh`/`powershell`/`cmd` on Windows, `bash`/`zsh`/`fish`/`$SHELL` elsewhere) and an `(envx)` prompt. `--shell CMD` runs a one-shot command instead. Both `run` and `shell` set `ENVX_ACTIVE=1`. |
| `copy` | Decrypts values and writes them to the system clipboard without printing to the terminal. |
| `key rotate` | Generates a fresh key and re-encrypts every secret, backing up the old one. `key export`/`key import` back up and restore a key file. |
| `web` | Local GUI at `http://127.0.0.1:4319` — a single embedded page backed by a JSON API, auto-refreshing every 2s. No daemon, localhost only. |
| `version` / `help` / `completions` | Version info, per-command help, and shell completion scripts. |

### Encryption backends

`config set encryption.key_backend keyring` (or `init --backend keyring`) keeps the encryption key in the **OS Credential Manager** (Windows) under `envx:<project root>` instead of `.envx/key.bin`. envx migrates the existing key automatically, in **either** direction.

---

## 🔄 Migrating from `.env`

```bash
envx init
envx import .env                    # or: envx set API_KEY --secret (prompts, no echo)
envx run -- python app.py           # add --overlay while the legacy .env still exists
```

`run --overlay` merges a legacy `.env` at the project root as a stopgap; values managed by envx always win. Once everything is migrated, `envx import .env`, delete the old `.env`, and turn the setting back off.

---

## 🗂 Storage

```
.envx/config.json          versioned config (schema v2)
.envx/config.backup.json   previous config before each write
.envx/config.lock          write lock
.envx/key.bin              encryption key, file backend (gitignored)
.envx/key.old.bin          previous key after rotation, file backend (gitignored)
```

With `encryption.key_backend=keyring`, no key files are written — the key lives in the OS Credential Manager.

Encrypted values are stored as `enc:<fernet-token>` and redacted by default in `envx list`. Use `--show-secrets` to reveal them.

---

## 🧩 Compatibility

The Go binary reads configs created by the earlier Python prototype (legacy schema and `key.key`), migrates them to the current schema in memory, and decrypts existing secrets using the same Fernet wire format. The Python reference implementation lives in `legacy-python/`.

---

## 🏗 Development

```bash
go test ./...      # run all tests
go vet ./...       # static analysis
gofmt -l .         # formatting check

cd npm && npm install   # smoke-test the npm package installer
```

Contributions welcome — see [CONTRIBUTING.md](CONTRIBUTING.md).

---

## 📄 License

[MIT](LICENSE) © 2026 TJ-programmer
