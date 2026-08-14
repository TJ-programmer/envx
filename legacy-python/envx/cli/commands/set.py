"""Set command."""

from __future__ import annotations

import typer

from envx.cli.commands.common import build_service, ensure_single_mode, handle_error
from envx.core.exceptions import EnvxError


def register(app: typer.Typer) -> None:
    """Register the set command."""

    @app.command(help="Set or update an environment variable.")
    def set(
        key: str = typer.Argument(..., help="Environment variable name."),
        value: str = typer.Argument(..., help="Environment variable value."),
        env: str | None = typer.Option(
            None, "--env", help="Target environment. Defaults to the active environment."
        ),
        secret: bool = typer.Option(
            False, "--secret", help="Encrypt and store the value as a secret."
        ),
        plain: bool = typer.Option(
            False, "--plain", help="Store the value without encryption."
        ),
    ) -> None:
        ensure_single_mode(secret, plain)
        service = build_service()
        try:
            service.set_variable(key=key, value=value, env_name=env, secret=secret)
        except EnvxError as exc:
            handle_error(exc)
        typer.echo(f"Stored '{key}' in environment '{env or service.load_config().active_env}'.")
