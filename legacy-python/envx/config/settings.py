"""Runtime settings for envx."""

from __future__ import annotations

from dataclasses import dataclass


@dataclass(slots=True)
class LoggingSettings:
    """Logging configuration toggles."""

    verbose: bool = False
    quiet: bool = False
    json_logs: bool = False
