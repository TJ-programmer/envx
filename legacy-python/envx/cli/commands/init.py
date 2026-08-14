"""Init command."""

from __future__ import annotations

import typer

from envx.cli.commands.common import build_service, handle_error
from envx.core.exceptions import EnvxError


def register(app: typer.Typer) -> None:
    """Register the init command."""

    @app.command(help="Initialize envx for the current project.")
    def init(
        env: str = typer.Option("dev", "--env", help="Initial environment name."),
        force: bool = typer.Option(
            False, "--force", help="Replace an existing envx config for this project."
        ),
    ) -> None:
        service = build_service()
        try:
            config = service.init_project(env_name=env, force=force)
        except EnvxError as exc:
            handle_error(exc)
        typer.echo(
            f"Initialized envx in '.envx/' with active environment '{config.active_env}'."
        )
