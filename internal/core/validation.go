package core

import (
	"fmt"
	"regexp"

	"envx/internal/errs"
)

var (
	envNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,63}$`)
	varNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
)

func ValidateEnvName(name string) error {
	if name == "" || !envNamePattern.MatchString(name) {
		return fmt.Errorf("%w: environment name must start with a letter or number and only contain letters, numbers, '.', '_' or '-'", errs.ErrVariableValidation)
	}
	return nil
}

func ValidateVariableName(name string) error {
	if name == "" || !varNamePattern.MatchString(name) {
		return fmt.Errorf("%w: variable name must start with a letter or underscore and contain only letters, numbers, and underscores", errs.ErrVariableValidation)
	}
	return nil
}

func ValidateConfig(cfg *ConfigFile) error {
	if cfg.Version != SchemaVersion {
		return fmt.Errorf("%w: config schema version %d, expected %d", errs.ErrConfigVersion, cfg.Version, SchemaVersion)
	}
	if err := ValidateEnvName(cfg.ActiveEnv); err != nil {
		return err
	}
	if _, ok := cfg.Environments[cfg.ActiveEnv]; !ok {
		return fmt.Errorf("%w: active environment does not exist in the config", errs.ErrVariableValidation)
	}
	for name, environment := range cfg.Environments {
		if err := ValidateEnvName(name); err != nil {
			return err
		}
		if environment.Name != name {
			return fmt.Errorf("%w: environment name mismatch in config", errs.ErrVariableValidation)
		}
		for variableName := range environment.Variables {
			if err := ValidateVariableName(variableName); err != nil {
				return err
			}
		}
	}
	return nil
}
