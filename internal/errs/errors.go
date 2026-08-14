package errs

import "errors"

var (
	ErrProjectNotInitialized = errors.New("this project is not initialized. Run 'envx init' first.")
	ErrAlreadyInitialized    = errors.New("project is already initialized. Use '--force' to replace the config.")
	ErrConfigCorrupted       = errors.New("config file is corrupted")
	ErrConfigVersion         = errors.New("unsupported config schema version")
	ErrEnvironmentNotFound   = errors.New("environment not found")
	ErrEnvironmentConflict   = errors.New("environment operation violates constraints")
	ErrVariableValidation    = errors.New("invalid value")
	ErrSecretKey             = errors.New("encryption key missing or invalid")
	ErrCommandExecution      = errors.New("command execution failed")
)
