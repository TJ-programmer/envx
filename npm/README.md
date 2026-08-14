# envx

**A faster, safer drop-in replacement for `.env` files.**

Project-local environment variables. Secrets encrypted at rest. Zero daemons, zero network, zero third-party.

This is the npm wrapper for the [envx](https://github.com/TJ-programmer/envx) single static binary. The package installs as the `envx` command. The prebuilt binary for your platform is downloaded from the GitHub release and verified against its SHA-256 checksum — no compiler, runtime, or interpreter needed.

The binary is fetched during `postinstall` when a package manager runs lifecycle scripts (npm). If scripts are skipped or blocked — pnpm ≥10 and bun block them by default — `cli.js` lazily downloads the binary on first run instead, so `envx` still just works on every manager.

Works identically with **npm**, **pnpm**, and **bun**.

## Install

```bash
npm install -g @tj-programmer/envx
pnpm add -g @tj-programmer/envx
bun add -g @tj-programmer/envx
```

Or run it on demand without installing:

```bash
npx @tj-programmer/envx --version
```

On pnpm ≥10 / bun the first `envx` invocation downloads the binary into the user cache (`%LOCALAPPDATA%\envx\bin\` on Windows, `~/.cache/envx/bin/` elsewhere) and verifies its checksum.

## Quickstart

```bash
envx init
envx set PORT 8000
envx set API_KEY my-secret --secret
envx run -- python app.py
```

## Overrides

- `ENVX_VERSION` — install a specific version instead of the package version.
- `ENVX_REPO` — GitHub `owner/repo` to fetch from (for forks; default `TJ-programmer/envx`).
- `ENVX_DOWNLOAD_BASE` — release download base URL (default `https://github.com/$REPO/releases/download`).

## Docs

Full documentation, commands, and storage layout: [github.com/TJ-programmer/envx](https://github.com/TJ-programmer/envx)

## License

[MIT](https://github.com/TJ-programmer/envx/blob/main/LICENSE)
