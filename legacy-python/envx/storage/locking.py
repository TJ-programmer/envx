"""Simple file locking helpers for local project storage."""

from __future__ import annotations

import time
from pathlib import Path


class FileLock:
    """A lock file using atomic creation."""

    def __init__(self, path: Path, timeout: float = 5.0, poll_interval: float = 0.1) -> None:
        self.path = path
        self.timeout = timeout
        self.poll_interval = poll_interval
        self._locked = False

    def __enter__(self) -> "FileLock":
        deadline = time.monotonic() + self.timeout
        while time.monotonic() < deadline:
            try:
                handle = self.path.open("x", encoding="utf-8")
            except FileExistsError:
                time.sleep(self.poll_interval)
                continue

            handle.write(str(time.time()))
            handle.close()
            self._locked = True
            return self

        raise TimeoutError(f"Timed out waiting for lock at {self.path}.")

    def __exit__(self, exc_type: object, exc: object, tb: object) -> None:
        if self._locked and self.path.exists():
            self.path.unlink(missing_ok=True)
        self._locked = False
