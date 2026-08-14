"""Validation helpers for config and inputs."""

from __future__ import annotations

import re

from envx.core.exceptions import ConfigVersionError, VariableValidationError
from envx.core.models import ConfigFile, SCHEMA_VERSION

ENV_NAME_PATTERN = re.compile(r"^[A-Za-z0-9][A-Za-z0-9_.-]{0,63}$")
VARIABLE_NAME_PATTERN = re.compile(r"^[A-Za-z_][A-Za-z0-9_]*$")


def validate_env_name(name: str) -> str:
    """Validate an environment name."""
    if not name or not ENV_NAME_PATTERN.fullmatch(name):
        raise VariableValidationError(
            "Environment name must start with a letter or number and only contain letters, numbers, '.', '_' or '-'."
        )
    return name


def validate_variable_name(name: str) -> str:
    """Validate an environment variable name."""
    if not name or not VARIABLE_NAME_PATTERN.fullmatch(name):
        raise VariableValidationError(
            "Variable name must start with a letter or underscore and contain only letters, numbers, and underscores."
        )
    return name


def validate_variable_value(value: str) -> str:
    """Validate a variable value."""
    if value is None:
        raise VariableValidationError("Variable value cannot be null.")
    return value


def validate_config(config: ConfigFile) -> ConfigFile:
    """Validate the full config model."""
    if config.version != SCHEMA_VERSION:
        raise ConfigVersionError(
            f"Unsupported config schema version {config.version}. Expected {SCHEMA_VERSION}."
        )

    validate_env_name(config.active_env)
    if config.active_env not in config.environments:
        raise VariableValidationError("Active environment does not exist in the config.")

    for env_name, environment in config.environments.items():
        validate_env_name(env_name)
        if environment.name != env_name:
            raise VariableValidationError("Environment name mismatch in config.")
        for variable_name, entry in environment.variables.items():
            validate_variable_name(variable_name)
            validate_variable_value(entry.value)

    return config
