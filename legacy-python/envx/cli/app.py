"""Typer application for envx."""

from __future__ import annotations

import typer

from envx.cli.commands import env as env_command
from envx.cli.commands import init as init_command
from envx.cli.commands import list as list_command
from envx.cli.commands import run as run_command
from envx.cli.commands import set as set_command
from envx.cli.commands.common import configure_cli_logging

app = typer.Typer(
    help="Project-local environment variable manager with secret storage and subprocess injection."
)


@app.callback()
def main(
    verbose: bool = typer.Option(False, "--verbose", help="Enable verbose logging."),
    quiet: bool = typer.Option(False, "--quiet", help="Only show errors."),
    json_logs: bool = typer.Option(
        False, "--json-logs", help="Emit structured JSON logs to stderr."
    ),
) -> None:
    """Configure shared CLI state."""
    configure_cli_logging(verbose=verbose, quiet=quiet, json_logs=json_logs)


init_command.register(app)
set_command.register(app)
list_command.register(app)
run_command.register(app)
env_command.register(app)
