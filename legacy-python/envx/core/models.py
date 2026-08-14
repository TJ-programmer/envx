"""Typed data models for envx state."""

from __future__ import annotations

from dataclasses import dataclass, field
from datetime import datetime, timezone


SCHEMA_VERSION = 1


def utc_now() -> str:
    """Return an ISO 8601 UTC timestamp."""
    return datetime.now(timezone.utc).replace(microsecond=0).isoformat()


@dataclass(slots=True)
class VariableEntry:
    """A single environment variable definition."""

    value: str
    is_secret: bool = False
    created_at: str = field(default_factory=utc_now)
    updated_at: str = field(default_factory=utc_now)

    def to_dict(self) -> dict[str, object]:
        return {
            "value": self.value,
            "is_secret": self.is_secret,
            "created_at": self.created_at,
            "updated_at": self.updated_at,
        }

    @classmethod
    def from_dict(cls, data: dict[str, object]) -> "VariableEntry":
        return cls(
            value=str(data["value"]),
            is_secret=bool(data.get("is_secret", False)),
            created_at=str(data.get("created_at", utc_now())),
            updated_at=str(data.get("updated_at", utc_now())),
        )


@dataclass(slots=True)
class EnvironmentSpec:
    """A named environment with variables."""

    name: str
    variables: dict[str, VariableEntry] = field(default_factory=dict)

    def to_dict(self) -> dict[str, object]:
        return {
            "name": self.name,
            "variables": {key: entry.to_dict() for key, entry in self.variables.items()},
        }

    @classmethod
    def from_dict(cls, data: dict[str, object]) -> "EnvironmentSpec":
        raw_variables = data.get("variables", {})
        if not isinstance(raw_variables, dict):
            raise TypeError("Environment variables must be a mapping.")

        return cls(
            name=str(data["name"]),
            variables={
                key: VariableEntry.from_dict(value)
                for key, value in raw_variables.items()
                if isinstance(value, dict)
            },
        )


@dataclass(slots=True)
class ConfigFile:
    """Root configuration model."""

    version: int = SCHEMA_VERSION
    active_env: str = "dev"
    environments: dict[str, EnvironmentSpec] = field(default_factory=dict)
    metadata: dict[str, object] = field(default_factory=dict)

    def to_dict(self) -> dict[str, object]:
        return {
            "version": self.version,
            "active_env": self.active_env,
            "environments": {
                key: environment.to_dict() for key, environment in self.environments.items()
            },
            "metadata": self.metadata,
        }

    @classmethod
    def create_default(cls, env_name: str = "dev") -> "ConfigFile":
        return cls(
            version=SCHEMA_VERSION,
            active_env=env_name,
            environments={env_name: EnvironmentSpec(name=env_name)},
            metadata={"created_at": utc_now()},
        )

    @classmethod
    def from_dict(cls, data: dict[str, object]) -> "ConfigFile":
        raw_environments = data.get("environments", {})
        if not isinstance(raw_environments, dict):
            raise TypeError("Environments must be a mapping.")

        environments: dict[str, EnvironmentSpec] = {}
        for key, raw_environment in raw_environments.items():
            if not isinstance(raw_environment, dict):
                raise TypeError("Environment entry must be an object.")
            merged = {"name": key, **raw_environment}
            environments[key] = EnvironmentSpec.from_dict(merged)

        return cls(
            version=int(data["version"]),
            active_env=str(data["active_env"]),
            environments=environments,
            metadata=dict(data.get("metadata", {})),
        )
