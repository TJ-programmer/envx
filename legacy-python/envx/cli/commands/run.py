"""Run command."""

from __future__ import annotations

import typer

from envx.cli.commands.common import build_service, handle_error
from envx.core.exceptions import EnvxError
from envx.runtime.command import build_command_spec


RUN_HELP = """Run a command with envx-injected variables.

Examples:
  envx run -- python app.py --port 8000
  envx run -- uvicorn api:app --reload
  envx run -- docker run --rm my-image
  envx run --shell "uvicorn api:app --reload | tee app.log"
"""


def register(app: typer.Typer) -> None:
    """Register the run command."""

    @app.command(
        help=RUN_HELP,
        context_settings={"allow_extra_args": True, "ignore_unknown_options": True},
    )
    def run(
        ctx: typer.Context,
        env: str | None = typer.Option(
            None, "--env", help="Target environment. Defaults to the active environment."
        ),
        shell_command: str | None = typer.Option(
            None,
            "--shell",
            help="Run through the system shell. Required for pipes, redirects, or shell operators.",
        ),
    ) -> None:
        service = build_service()
        command = build_command_spec(ctx.args, shell_command)
        try:
            exit_code = service.run_command(command=command, env_name=env)
        except EnvxError as exc:
            handle_error(exc)
        raise typer.Exit(code=exit_code)
