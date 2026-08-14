"""Application services for envx workflows."""

from __future__ import annotations

from dataclasses import replace
from typing import Callable

from envx.config.paths import ProjectPaths
from envx.core.exceptions import EnvironmentConflictError, EnvironmentNotFoundError
from envx.core.models import ConfigFile, EnvironmentSpec, VariableEntry, utc_now
from envx.core.validation import (
    validate_env_name,
    validate_variable_name,
    validate_variable_value,
)
from envx.runtime.command import CommandSpec
from envx.runtime.injector import run_with_env
from envx.security.crypto import CryptoManager
from envx.security.redaction import redact_value
from envx.storage.json_store import JsonConfigStore

Hook = Callable[..., None]


class HookRegistry:
    """Minimal extension seams for future plugins."""

    def __init__(self) -> None:
        self.post_init_hooks: list[Hook] = []
        self.pre_run_hooks: list[Hook] = []

    def register_post_init(self, hook: Hook) -> None:
        self.post_init_hooks.append(hook)

    def register_pre_run(self, hook: Hook) -> None:
        self.pre_run_hooks.append(hook)


class Exporter:
    """Placeholder exporter interface for future integrations."""

    def export(self, config: ConfigFile) -> str:
        raise NotImplementedError


class EnvironmentResolver:
    """Resolves active environment names."""

    @staticmethod
    def resolve(config: ConfigFile, requested_env: str | None) -> str:
        env_name = requested_env or config.active_env
        if env_name not in config.environments:
            raise EnvironmentNotFoundError(f"Environment '{env_name}' was not found.")
        return env_name


class EnvxService:
    """Primary service layer for envx operations."""

    def __init__(
        self,
        store: JsonConfigStore,
        crypto: CryptoManager,
        paths: ProjectPaths,
        hooks: HookRegistry | None = None,
    ) -> None:
        self.store = store
        self.crypto = crypto
        self.paths = paths
        self.hooks = hooks or HookRegistry()

    def init_project(self, env_name: str = "dev", force: bool = False) -> ConfigFile:
        """Initialize a project-local config and key."""
        validate_env_name(env_name)
        self.crypto.ensure_key()
        config = ConfigFile.create_default(env_name=env_name)
        stored = self.store.initialize(config, force=force)
        for hook in self.hooks.post_init_hooks:
            hook(self.paths.root, stored)
        return stored

    def load_config(self) -> ConfigFile:
        """Load the current config."""
        return self.store.load()

    def set_variable(
        self,
        key: str,
        value: str,
        env_name: str | None = None,
        secret: bool = False,
    ) -> ConfigFile:
        """Set or update a variable in the target environment."""
        validate_variable_name(key)
        validate_variable_value(value)
        config = self.store.load()
        resolved_env = EnvironmentResolver.resolve(config, env_name)
        environment = config.environments[resolved_env]
        current = environment.variables.get(key)
        encrypted_value = self.crypto.encrypt(value) if secret else value

        entry = VariableEntry(
            value=encrypted_value,
            is_secret=secret,
            created_at=current.created_at if current else utc_now(),
            updated_at=utc_now(),
        )
        environment.variables[key] = entry
        self.store.save(config)
        return config

    def list_variables(
        self,
        env_name: str | None = None,
        show_secrets: bool = False,
    ) -> list[dict[str, str]]:
        """Return variables for display."""
        config = self.store.load()
        resolved_env = EnvironmentResolver.resolve(config, env_name)
        environment = config.environments[resolved_env]
        rows: list[dict[str, str]] = []
        for key in sorted(environment.variables):
            entry = environment.variables[key]
            display_value = (
                self.crypto.decrypt(entry.value)
                if show_secrets
                else redact_value(entry.value, entry.is_secret)
            )
            rows.append(
                {
                    "key": key,
                    "value": display_value,
                    "secret": str(entry.is_secret).lower(),
                    "environment": resolved_env,
                }
            )
        return rows

    def create_environment(self, env_name: str) -> ConfigFile:
        """Create a new environment."""
        validate_env_name(env_name)
        config = self.store.load()
        if env_name in config.environments:
            raise EnvironmentConflictError(f"Environment '{env_name}' already exists.")
        config.environments[env_name] = EnvironmentSpec(name=env_name)
        self.store.save(config)
        return config

    def use_environment(self, env_name: str) -> ConfigFile:
        """Mark an environment as active."""
        config = self.store.load()
        EnvironmentResolver.resolve(config, env_name)
        config.active_env = env_name
        self.store.save(config)
        return config

    def delete_environment(self, env_name: str) -> ConfigFile:
        """Delete an environment that is not active."""
        config = self.store.load()
        EnvironmentResolver.resolve(config, env_name)
        if env_name == config.active_env:
            raise EnvironmentConflictError("Cannot delete the active environment.")
        del config.environments[env_name]
        self.store.save(config)
        return config

    def list_environments(self) -> list[dict[str, str]]:
        """Return environments for display."""
        config = self.store.load()
        rows: list[dict[str, str]] = []
        for env_name in sorted(config.environments):
            rows.append(
                {
                    "name": env_name,
                    "active": str(env_name == config.active_env).lower(),
                    "variables": str(len(config.environments[env_name].variables)),
                }
            )
        return rows

    def resolve_runtime_env(self, env_name: str | None = None) -> tuple[str, dict[str, str]]:
        """Build decrypted env vars for a child process."""
        config = self.store.load()
        resolved_env = EnvironmentResolver.resolve(config, env_name)
        environment = config.environments[resolved_env]
        env_vars = {
            key: self.crypto.decrypt(entry.value) for key, entry in environment.variables.items()
        }
        return resolved_env, env_vars

    def run_command(self, command: CommandSpec, env_name: str | None = None) -> int:
        """Run a child process with injected environment variables."""
        resolved_env, env_vars = self.resolve_runtime_env(env_name)
        for hook in self.hooks.pre_run_hooks:
            hook(resolved_env, command, env_vars.copy())
        return run_with_env(command, env_vars)
