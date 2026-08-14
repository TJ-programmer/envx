"""List command."""

from __future__ import annotations

from typing import Annotated

import typer

from envx.cli.commands.common import build_service, handle_error, render_rows
from envx.core.exceptions import EnvxError


def register(app: typer.Typer) -> None:
    """Register the list command."""

    @app.command(name="list", help="List environment variables.")
    def list_command(
        env: str | None = typer.Option(
            None, "--env", help="Target environment. Defaults to the active environment."
        ),
        show_secrets: bool = typer.Option(
            False, "--show-secrets", help="Decrypt and display secret values."
        ),
        output_format: Annotated[
            str,
            typer.Option(
                "--format",
                help="Output format.",
                case_sensitive=False,
            ),
        ] = "table",
    ) -> None:
        service = build_service()
        try:
            rows = service.list_variables(env_name=env, show_secrets=show_secrets)
        except EnvxError as exc:
            handle_error(exc)
        render_rows(rows, output_format.lower())
