"""Structured logging configuration."""

from __future__ import annotations

import json
import logging
import sys

from envx.config.settings import LoggingSettings


class JsonFormatter(logging.Formatter):
    """Format log records as JSON."""

    def format(self, record: logging.LogRecord) -> str:
        payload = {
            "level": record.levelname,
            "message": record.getMessage(),
            "logger": record.name,
        }
        return json.dumps(payload)


def configure_logging(settings: LoggingSettings) -> None:
    """Configure root logging for CLI execution."""
    level = logging.INFO
    if settings.verbose:
        level = logging.DEBUG
    if settings.quiet:
        level = logging.ERROR

    handler = logging.StreamHandler(sys.stderr)
    if settings.json_logs:
        handler.setFormatter(JsonFormatter())
    else:
        handler.setFormatter(logging.Formatter("%(levelname)s: %(message)s"))

    root_logger = logging.getLogger()
    root_logger.handlers.clear()
    root_logger.setLevel(level)
    root_logger.addHandler(handler)
