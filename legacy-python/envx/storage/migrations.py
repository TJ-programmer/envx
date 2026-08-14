"""Config migration helpers."""

from __future__ import annotations

from envx.core.models import ConfigFile, SCHEMA_VERSION, VariableEntry


def migrate_config(raw: dict[str, object]) -> dict[str, object]:
    """Normalize legacy config shapes into the current schema."""
    if "version" in raw:
        return raw

    active_env = str(raw.get("active_env", "dev"))
    environments = raw.get("environments", {})
    migrated_environments: dict[str, object] = {}

    if isinstance(environments, dict):
        for env_name, variables in environments.items():
            migrated_variables: dict[str, object] = {}
            if isinstance(variables, dict):
                for variable_name, variable_value in variables.items():
                    is_secret = (
                        isinstance(variable_value, str) and variable_value.startswith("enc:")
                    )
                    migrated_variables[str(variable_name)] = VariableEntry(
                        value=str(variable_value),
                        is_secret=is_secret,
                    ).to_dict()

            migrated_environments[str(env_name)] = {
                "name": str(env_name),
                "variables": migrated_variables,
            }

    normalized = ConfigFile(
        version=SCHEMA_VERSION,
        active_env=active_env,
        environments={},
        metadata={"migrated_from_legacy": True},
    ).to_dict()
    normalized["environments"] = migrated_environments
    return normalized
