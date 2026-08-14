"""Application-specific exceptions."""


class EnvxError(Exception):
    """Base exception for envx."""


class ProjectNotInitializedError(EnvxError):
    """Raised when the current project has not been initialized."""


class ConfigCorruptedError(EnvxError):
    """Raised when configuration cannot be loaded safely."""


class ConfigVersionError(EnvxError):
    """Raised when the config schema version is unsupported."""


class EnvironmentNotFoundError(EnvxError):
    """Raised when the requested environment does not exist."""


class EnvironmentConflictError(EnvxError):
    """Raised when an environment operation would violate constraints."""


class VariableValidationError(EnvxError):
    """Raised when a variable name or value is invalid."""


class SecretKeyError(EnvxError):
    """Raised when the encryption key is missing or invalid."""


class CommandExecutionError(EnvxError):
    """Raised when a child command cannot be executed."""

