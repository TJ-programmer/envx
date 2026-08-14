"""Command parsing and execution models."""

from __future__ import annotations

from dataclasses import dataclass
from typing import Sequence


@dataclass(slots=True)
class CommandSpec:
    """A child process specification."""

    argv: list[str] | None = None
    shell_command: str | None = None

    @property
    def is_shell(self) -> bool:
        """Return whether the command is shell-based."""
        return self.shell_command is not None


def build_command_spec(args: Sequence[str], shell_command: str | None) -> CommandSpec:
    """Normalize CLI inputs into a command spec."""
    if shell_command:
        return CommandSpec(shell_command=shell_command)
    return CommandSpec(argv=list(args))
