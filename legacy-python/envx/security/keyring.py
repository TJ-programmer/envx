"""Local key file management."""

from __future__ import annotations

import os
from pathlib import Path

from cryptography.fernet import Fernet

from envx.config.paths import ProjectPaths
from envx.core.exceptions import SecretKeyError


def _restrict_permissions(path: Path) -> None:
    """Best-effort file permission hardening."""
    try:
        os.chmod(path, 0o600)
    except OSError:
        return


class LocalKeyProvider:
    """A project-local Fernet key provider."""

    def __init__(self, paths: ProjectPaths) -> None:
        self.paths = paths

    def ensure_key(self) -> bytes:
        """Return an existing key or create a new one."""
        self.paths.config_dir.mkdir(parents=True, exist_ok=True)

        if self.paths.key_path.exists():
            key = self.paths.key_path.read_bytes()
            self._validate_key(key)
            return key

        if self.paths.legacy_key_path.exists():
            key = self.paths.legacy_key_path.read_bytes()
            self._validate_key(key)
            self.paths.key_path.write_bytes(key)
            _restrict_permissions(self.paths.key_path)
            return key

        key = Fernet.generate_key()
        self.paths.key_path.write_bytes(key)
        _restrict_permissions(self.paths.key_path)
        return key

    def load_key(self) -> bytes:
        """Load an existing key without creating a new one."""
        if self.paths.key_path.exists():
            key = self.paths.key_path.read_bytes()
            self._validate_key(key)
            return key

        if self.paths.legacy_key_path.exists():
            key = self.paths.legacy_key_path.read_bytes()
            self._validate_key(key)
            return key

        raise SecretKeyError(
            "Encryption key not found. Run 'envx init' or restore '.envx/key.bin'."
        )

    @staticmethod
    def _validate_key(key: bytes) -> None:
        if len(key) != 44:
            raise SecretKeyError("Encryption key is invalid or corrupted.")
