# AGENTS.md

Guidance for AI agents and humans working on this repository.

## Project

`envx` is a single static Go binary that replaces `.env` files. It stores project-local environment variables in `<project>/.envx/`, injects them into commands/shells/scripts, and encrypts secrets at rest with a project-local Fernet key. Zero daemons, zero network, zero third-party at runtime.

- Module: `envx` (no external Go dependencies — `go.mod` is dependency-free).
- Go 1.26+. Binary entry point: `./cmd/envx`.
- Releases: GitHub Actions + GoReleaser cross-compile static binaries for linux/darwin/windows (amd64 + arm64) into `envx_<version>_<os>_<arch>.tar.gz|.zip` plus `checksums.txt`.

## Commands

```bash
go build -o bin/envx ./cmd/envx   # build
go test ./... -count=1            # run all tests (fast, no network/keyring/home)
go vet ./...                      # static analysis
gofmt -l .                        # must print NOTHING
cd npm && npm install             # smoke-test the npm package installer
```

Always run `go test`, `go vet`, and `gofmt -l .` after making Go changes. `gofmt -l .` must be empty.

## Layout

```
cmd/envx/                 CLI entry point (flag wiring, signal handling)
internal/buildinfo/       version metadata (default "0.5.0", overridable via -ldflags)
internal/cli/             command implementations + tests
internal/config/          versioned project config (schema v2)
internal/core/            service layer (init/set/get/run/shell/copy…)
internal/crypto/          Fernet encryption
internal/errs/            shared errors
internal/gitignore/       .gitignore management
internal/keyring/         OS keyring backend (Windows Credential Manager)
internal/run/             process execution (run/shell, Windows + POSIX)
internal/store/           key/value storage + environment overlays
internal/web/             embedded web UI + JSON API
legacy-python/            original Python prototype (reference only)
scripts/                  install.sh / install.ps1 (one-liner installers)
npm/                      npm/pnpm/bun package; install.js downloads the release binary
brew/envx.rb              Homebrew formula (builds from source; copy into a tap repo)
scoop/envx.json           Scoop manifest (prebuilt binaries; fill hashes on release)
.github/workflows/        release.yml (GoReleaser on tag push)
```

## Conventions

- No comments unless they add real value; let the code speak.
- Keep `go vet` clean and `gofmt`-formatted.
- Tests are deterministic and fast — never touch the network, the real keyring, or real `$HOME`.
- Process-level tests (run/shell) use a self-re-exec helper pattern: the test binary re-invokes itself as a child with a special env var (e.g. `ENVX_RUN_HELPER`) and asserts on output.
- Side-effect seams (clipboard, shell launch) are swappable exported `var`s (e.g. `clipboard.Write`) — swap them in tests instead of monkey-patching.
- Table-driven tests preferred. One logical change per PR. Conventional Commits style (feat:/fix:/docs:/test:/chore:/refactor:).

## Environment / version

`envx` checks `internal/buildinfo.Version`; embedded at build time via `-ldflags "-X envx/internal/buildinfo.Version=vX.Y.Z"`. Bump it to the next dev version after each release.

## Release process (maintainers)

1. Tag `vX.Y.Z` and push — release.yml + GoReleaser builds binaries + checksums.
2. Bump `internal/buildinfo.Version` to next dev version.
3. Sync distribution channels (keep `npm/package.json` version in sync):
   - **npm**: bump `version` in `npm/package.json`, run `npm publish` from `npm/` (scoped `@tj-programmer/envx`, public via `publishConfig`; installs the `envx` command). The `postinstall` (`install.js`) downloads the matching release binary + verifies its SHA-256 from `checksums.txt` — no binary is committed.
   - **Homebrew**: update `url`/`sha256` in `brew/envx.rb` (formula builds from source).
   - **Scoop**: bump `version` in `scoop/envx.json` and fill the two `hash` fields from `checksums.txt`.

Installers and the npm package honor `ENVX_REPO` (default `TJ-programmer/envx`) and `ENVX_DOWNLOAD_BASE`; the npm package also honors `ENVX_VERSION`.

## Gotchas

- Windows archives are `.zip`; npm `install.js` extracts with `tar -xzf` (Windows ships bsdtar which reads zip). Linux/macOS are `.tar.gz` with GNU/BSD tar.
- GoReleaser archive names embed the version without the `v` prefix: `envx_0.5.0_linux_amd64.tar.gz`.
- `vendor/` is gitignored (used at npm install time; never commit downloaded binaries).
- `bin/` is gitignored — the committed tree must not contain build artifacts.
