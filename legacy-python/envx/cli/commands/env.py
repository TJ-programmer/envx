"""Environment lifecycle commands."""

from __future__ import annotations

import typer

from envx.cli.commands.common import build_service, handle_error, render_rows
from envx.core.exceptions import EnvxError


def register(app: typer.Typer) -> None:
    """Register environment lifecycle commands."""
    env_app = typer.Typer(help="Manage named environments.")

    @env_app.command("create", help="Create a new environment.")
    def create(name: str = typer.Argument(..., help="Environment name.")) -> None:
        service = build_service()
        try:
            service.create_environment(name)
        except EnvxError as exc:
            handle_error(exc)
        typer.echo(f"Created environment '{name}'.")

    @env_app.command("use", help="Set the active environment.")
    def use(name: str = typer.Argument(..., help="Environment name.")) -> None:
        service = build_service()
        try:
            service.use_environment(name)
        except EnvxError as exc:
            handle_error(exc)
        typer.echo(f"Active environment is now '{name}'.")

    @env_app.command("delete", help="Delete an environment.")
    def delete(name: str = typer.Argument(..., help="Environment name.")) -> None:
        service = build_service()
        try:
            service.delete_environment(name)
        except EnvxError as exc:
            handle_error(exc)
        typer.echo(f"Deleted environment '{name}'.")

    @env_app.command("list", help="List available environments.")
    def list_envs() -> None:
        service = build_service()
        try:
            rows = service.list_environments()
        except EnvxError as exc:
            handle_error(exc)
        render_rows(rows, "table")

    app.add_typer(env_app, name="env")
