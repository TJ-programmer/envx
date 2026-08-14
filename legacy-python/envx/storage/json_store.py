"""JSON-backed config storage with migrations and atomic writes."""

from __future__ import annotations

import json
from pathlib import Path

from envx.config.paths import ProjectPaths
from envx.core.exceptions import ConfigCorruptedError, ProjectNotInitializedError
from envx.core.models import ConfigFile
from envx.core.validation import validate_config
from envx.storage.locking import FileLock
from envx.storage.migrations import migrate_config


class JsonConfigStore:
    """Load and persist envx config state."""

    def __init__(self, paths: ProjectPaths) -> None:
        self.paths = paths

    def exists(self) -> bool:
        """Return whether the config exists."""
        return self.paths.config_path.exists()

    def initialize(self, config: ConfigFile, force: bool = False) -> ConfigFile:
        """Create a fresh config file."""
        if self.exists() and not force:
            raise ConfigCorruptedError(
                "Project is already initialized. Use '--force' to replace the config."
            )

        self.save(config)
        return config

    def load(self) -> ConfigFile:
        """Load and validate the current config."""
        if not self.paths.config_path.exists():
            raise ProjectNotInitializedError(
                "This project is not initialized. Run 'envx init' first."
            )

        raw = self._load_json(self.paths.config_path)
        normalized = migrate_config(raw)
        try:
            config = ConfigFile.from_dict(normalized)
        except (KeyError, TypeError, ValueError) as exc:
            raise ConfigCorruptedError(
                "Config file is corrupted. Restore '.envx/config.backup.json' or re-run 'envx init --force'."
            ) from exc

        return validate_config(config)

    def save(self, config: ConfigFile) -> None:
        """Atomically persist the current config."""
        validate_config(config)
        self.paths.config_dir.mkdir(parents=True, exist_ok=True)

        with FileLock(self.paths.lock_path):
            if self.paths.config_path.exists():
                self.paths.backup_path.write_text(
                    self.paths.config_path.read_text(encoding="utf-8"), encoding="utf-8"
                )

            temp_path = self.paths.config_path.with_suffix(".json.tmp")
            temp_path.write_text(
                json.dumps(config.to_dict(), indent=2, sort_keys=True) + "\n",
                encoding="utf-8",
            )
            temp_path.replace(self.paths.config_path)

    @staticmethod
    def _load_json(path: Path) -> dict[str, object]:
        try:
            return json.loads(path.read_text(encoding="utf-8"))
        except json.JSONDecodeError as exc:
            raise ConfigCorruptedError(
                "Config file contains invalid JSON. Restore '.envx/config.backup.json' or re-run 'envx init --force'."
            ) from exc
