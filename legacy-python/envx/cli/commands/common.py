"""Shared CLI helpers."""

from __future__ import annotations

import json
from pathlib import Path
from typing import Any

import typer

from envx.config.paths import resolve_project_paths
from envx.core.exceptions import EnvxError
from envx.core.services import EnvxService
from envx.observability.logging import configure_logging
from envx.config.settings import LoggingSettings
from envx.security.crypto import CryptoManager
from envx.security.keyring import LocalKeyProvider
from envx.storage.json_store import JsonConfigStore


def build_service(base_dir: Path | None = None) -> EnvxService:
    """Construct the service graph for the current project."""
    paths = resolve_project_paths(base_dir)
    store = JsonConfigStore(paths)
    crypto = CryptoManager(LocalKeyProvider(paths))
    return EnvxService(store=store, crypto=crypto, paths=paths)


def configure_cli_logging(verbose: bool, quiet: bool, json_logs: bool) -> None:
    """Configure root logging for the CLI."""
    configure_logging(
        LoggingSettings(verbose=verbose, quiet=quiet, json_logs=json_logs)
    )


def render_rows(rows: list[dict[str, str]], output_format: str) -> None:
    """Render rows in either table or JSON form."""
    if output_format == "json":
        typer.echo(json.dumps(rows, indent=2))
        return

    if not rows:
        typer.echo("No values found.")
        return

    columns = list(rows[0].keys())
    widths = {
        column: max(len(column), *(len(row[column]) for row in rows))
        for column in columns
    }
    typer.echo("  ".join(column.upper().ljust(widths[column]) for column in columns))
    for row in rows:
        typer.echo("  ".join(row[column].ljust(widths[column]) for column in columns))


def handle_error(exc: EnvxError) -> None:
    """Render a consistent user-facing error and exit."""
    typer.secho(f"Error: {exc}", fg=typer.colors.RED, err=True)
    raise typer.Exit(code=1)


def ensure_single_mode(secret: bool, plain: bool) -> None:
    """Validate mutually exclusive secret flags."""
    if secret and plain:
        raise typer.BadParameter("Use either '--secret' or '--plain', not both.")


def noop(*_: Any) -> None:
    """Default hook placeholder."""
    return
