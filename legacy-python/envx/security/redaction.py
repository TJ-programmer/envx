"""Secret redaction helpers."""

from __future__ import annotations


REDACTED_VALUE = "********"


def redact_value(value: str, is_secret: bool) -> str:
    """Return a safe display value."""
    if is_secret:
        return REDACTED_VALUE
    return value
