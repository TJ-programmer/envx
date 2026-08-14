"""Path helpers for project-local envx data."""

from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path


LEGACY_KEY_FILENAME = "key.key"


@dataclass(frozen=True, slots=True)
class ProjectPaths:
    """Resolved filesystem paths for envx storage."""

    root: Path
    config_dir: Path
    config_path: Path
    backup_path: Path
    lock_path: Path
    key_path: Path
    legacy_key_path: Path


def resolve_project_paths(base_dir: Path | None = None) -> ProjectPaths:
    """Resolve the current project's envx paths."""
    root = (base_dir or Path.cwd()).resolve()
    config_dir = root / ".envx"
    return ProjectPaths(
        root=root,
        config_dir=config_dir,
        config_path=config_dir / "config.json",
        backup_path=config_dir / "config.backup.json",
        lock_path=config_dir / "config.lock",
        key_path=config_dir / "key.bin",
        legacy_key_path=config_dir / LEGACY_KEY_FILENAME,
    )
