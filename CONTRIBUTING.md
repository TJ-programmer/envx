# Contributing to envx

Thanks for helping make envx better! 🎉 Whether you're fixing a typo, reporting a bug, or shipping a new feature — everything here is open and appreciated.

## Table of contents

- [Development setup](#development-setup)
- [Project layout](#project-layout)
- [Making changes](#making-changes)
- [Testing](#testing)
- [Code style](#code-style)
- [Commit conventions](#commit-conventions)
- [Opening a PR](#opening-a-pr)
- [Reporting issues](#reporting-issues)
- [Releasing a new version](#releasing-a-new-version)

## Development setup

Prerequisites:

- **Go 1.26+** ([download](https://go.dev/dl/))
- **git**

Clone and build:

```bash
git clone https://github.com/TJ-programmer/envx
cd envx
go build -o bin/envx ./cmd/envx
```

You now have a working `bin/envx` you can try in any scratch project:

```bash
envx init
envx set PORT 8000
envx set API_KEY test-secret --secret
envx run -- echo "$PORT"
```

## Project layout

```
cmd/envx/                 CLI entry point (flag wiring, signal handling)
internal/buildinfo/       version metadata (overridable via -ldflags)
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
scripts/                  install.sh / install.ps1 (release installers)
```

## Making changes

1. Fork the repo and clone your fork.
2. Create a feature branch:

   ```bash
   git checkout -b feat/my-change
   ```

3. Make your change, following the [code style](#code-style) and [testing](#testing) guidance below.
4. Run the full check suite (below) until everything passes.
5. Commit with a clear message (see [commit conventions](#commit-conventions)) and open a PR.

## Testing

Always run the full suite before pushing:

```bash
go test ./... -count=1
go vet ./...
gofmt -l .
```

`gofmt -l .` must print **nothing**.

### Writing tests

- Tests live next to the code: `internal/cli/phase6_test.go`, `internal/run/run_test.go`, `internal/core/shell_test.go`, etc.
- Keep tests **deterministic and fast** — they must not touch the network, the real keyring, or your real `$HOME`.
- For process-level tests (run/shell), the codebase uses a self-re-exec helper pattern: the test binary re-invokes itself as a child with a special env var (e.g. `ENVX_RUN_HELPER`, `ENVX_SHELL_HELPER`), and the parent asserts on the child's output.
- For side-effect seams (clipboard, shell launch), swap an exported `var` in the relevant package (e.g. `clipboard.Write`) rather than monkey-patching globals.

## Code style

- **No comments unless they add real value** — let the code speak.
- Run `gofmt` on everything you touch; keep `go vet` clean.
- Follow the patterns already in the codebase (error handling, output helpers, flag naming).
- Prefer table-driven tests where it fits.
- Keep changes focused: one logical change per PR.

## Commit conventions

We use lightweight [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add <what>
fix: correct <what>
docs: update <what>
test: cover <what>
chore: <housekeeping>
refactor: <what and why>
```

Examples from the repo history:

```
feat: Phase 5 - run --watch (auto-restart on change) + OS keyring backend
feat: adopt bnv improvements - envx shell, envx copy, ENVX_ACTIVE=1, ...
```

## Opening a PR

- Point the PR at `main`.
- Describe **what** changed and **why** — screenshots or terminal transcripts are great for CLI/UI changes.
- Link any related issue with `Closes #123`.
- Make sure CI (if enabled) is green; locally that means `go test ./...`, `go vet ./...`, and `gofmt -l .` all pass.

## Reporting issues

Before opening an issue:

1. Search existing issues to avoid duplicates.
2. Include your `envx --version` output.
3. Include a minimal reproduction — the commands you ran and the output you saw.

Feature requests are welcome too — describe the problem you're solving, not just the feature you want.

## Releasing a new version

Releases are automated via GitHub Actions + GoReleaser. Maintainers only:

1. Merge your changes to `main`.
2. Tag the release and push:

   ```bash
   git tag vX.Y.Z
   git push origin vX.Y.Z
   ```

3. The workflow cross-compiles static binaries for Linux/macOS/Windows (amd64 + arm64), attaches them and `checksums.txt` to a GitHub Release, and embeds the tag version via `-ldflags`.
4. Bump `internal/buildinfo.Version` on `main` to the next dev version (e.g. after releasing `v0.6.0`, set it to `0.6.1`).

> Test the installers after a release: the `scripts/install.sh` and `scripts/install.ps1` one-liners in the README should work from a clean machine.
