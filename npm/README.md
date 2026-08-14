# envx

**A faster, safer drop-in replacement for `.env` files.**

Project-local environment variables. Secrets encrypted at rest. Zero daemons, zero network, zero third-party.

This is an npm wrapper for the [envx](https://github.com/TJ-programmer/envx) single static binary. The package downloads the correct prebuilt binary for your platform from the GitHub release during install — no compiler, runtime, or interpreter needed.

Works identically with **npm**, **pnpm**, and **bun**.

## Install

```bash
npm install -g envx
pnpm add -g envx
bun add -g envx
```

Or run it on demand without installing:

```bash
npx envx --version
```

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
