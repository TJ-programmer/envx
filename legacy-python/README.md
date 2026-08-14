# envx

`envx` is a project-local environment variable manager for developers who want a small CLI with safer defaults than ad hoc `.env` files. It stores configuration under `.envx/`, supports multiple named environments, encrypts secrets at rest with a local key file, and injects variables into subprocesses without mutating your parent shell.

## Why envx

- Keep project-specific variables out of global shell state.
- Store secrets encrypted at rest instead of plain JSON values.
- Switch between `dev`, `staging`, and `prod` style environments from the same CLI.
- Run real commands such as `python`, `uvicorn`, and `docker` with injected variables.

## Install

```bash
pip install envx
```

For development:

```bash
pip install -e ".[dev]"
```

## Quickstart

```bash
envx init
envx set PORT 8000
envx set API_KEY my-secret --secret
envx list
envx run -- python app.py
```

## Command guide

Initialize a project:

```bash
envx init --env dev
```

Set plain or secret values:

```bash
envx set PORT 8000
envx set API_KEY my-secret --secret
envx set LOG_LEVEL debug --env staging
```

List values:

```bash
envx list
envx list --show-secrets
envx list --format json
```

Manage environments:

```bash
envx env create staging
envx env use staging
envx env list
envx env delete staging
```

Run commands with injected variables:

```bash
envx run -- python app.py --port 8000
envx run -- uvicorn api:app --reload
envx run -- docker run --rm my-image
envx run --shell "uvicorn api:app --reload | tee app.log"
```

## Storage layout

`envx` stores project data in:

- `.envx/config.json`
- `.envx/config.backup.json`
- `.envx/key.bin`

The config file is versioned and written atomically. Existing legacy configs are migrated when loaded and will be written back in the new structured schema.

## Secret model

Secret values are only encrypted when you opt in with `--secret`. By default:

- ciphertext is stored in `config.json`
- `envx list` redacts secret values
- `envx list --show-secrets` reveals them
- logs and error messages do not print secret values

This local key-file model protects secrets at rest from casual disclosure, but it does not protect against a fully compromised user account or machine.

## CI usage

`envx` is designed for local development and non-interactive runners:

```bash
envx init --force
envx set BUILD_ENV ci
envx run -- python -m pytest
```

Use `--json-logs` if your CI pipeline prefers structured stderr output.

## Developer onboarding

```bash
python -m venv venv
venv\Scripts\activate
pip install -e ".[dev]"
pytest
```

## Migration notes

Older `envx` prototypes stored:

- raw key/value maps in `.envx/config.json`
- a Fernet key in `.envx/key.key`

The current release automatically reads those legacy files, migrates the config shape in memory, and uses `.envx/key.bin` as the preferred key path going forward.
