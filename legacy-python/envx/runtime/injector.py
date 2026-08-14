"""Environment injection for subprocess execution."""

from __future__ import annotations

import os
import subprocess

from envx.core.exceptions import CommandExecutionError
from envx.runtime.command import CommandSpec


def run_with_env(command: CommandSpec, env_vars: dict[str, str]) -> int:
    """Execute a command with injected environment variables."""
    child_env = os.environ.copy()
    child_env.update(env_vars)

    try:
        if command.is_shell:
            completed = subprocess.run(command.shell_command, env=child_env, shell=True, check=False)
        else:
            if not command.argv:
                raise CommandExecutionError("No command was provided. Use 'envx run -- <command>'.")
            completed = subprocess.run(command.argv, env=child_env, shell=False, check=False)
    except FileNotFoundError as exc:
        missing = command.argv[0] if command.argv else "command"
        raise CommandExecutionError(
            f"Command not found: {missing}. Use '--shell' only for shell syntax such as pipes."
        ) from exc
    except OSError as exc:
        raise CommandExecutionError(f"Failed to execute child process: {exc}.") from exc

    return completed.returncode
