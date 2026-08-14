package core

import (
	"fmt"
	"sort"
	"strconv"

	"envx/internal/crypto"
	"envx/internal/errs"
)

type ConfigStore interface {
	Exists() bool
	Initialize(cfg *ConfigFile, force bool) error
	Load() (*ConfigFile, error)
	Save(cfg *ConfigFile) error
}

type Crypter interface {
	EnsureKey() error
	Encrypt(value string) (string, error)
	Decrypt(value string) (string, error)
}

type CommandRunner interface {
	Run(argv []string, shellCmd string, envVars map[string]string) (int, error)
}

type VariableRow struct {
	Key         string
	Value       string
	Secret      string
	Environment string
}

type EnvironmentRow struct {
	Name      string
	Active    string
	Variables string
}

type EnvxService struct {
	store  ConfigStore
	crypto Crypter
	runner CommandRunner
}

func NewService(store ConfigStore, crypto Crypter, runner CommandRunner) *EnvxService {
	return &EnvxService{store: store, crypto: crypto, runner: runner}
}

func (s *EnvxService) InitProject(envName string, force bool) (*ConfigFile, error) {
	if err := ValidateEnvName(envName); err != nil {
		return nil, err
	}
	if err := s.crypto.EnsureKey(); err != nil {
		return nil, err
	}
	cfg := NewConfigFile(envName)
	if err := s.store.Initialize(cfg, force); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (s *EnvxService) Load() (*ConfigFile, error) {
	return s.store.Load()
}

func (s *EnvxService) SetVariable(key, value, envName string, secret, plain bool) (*ConfigFile, error) {
	if err := ValidateVariableName(key); err != nil {
		return nil, err
	}
	cfg, err := s.store.Load()
	if err != nil {
		return nil, err
	}
	resolved, err := resolveEnvironment(cfg, envName)
	if err != nil {
		return nil, err
	}

	encrypt := secret
	if !encrypt && !plain && cfg.Encryption.DefaultEncrypt {
		encrypt = true
	}

	storedValue := value
	if encrypt {
		storedValue, err = s.crypto.Encrypt(value)
		if err != nil {
			return nil, err
		}
	}

	environment := cfg.Environments[resolved]
	if environment.Variables == nil {
		environment.Variables = map[string]VariableEntry{}
	}
	current := environment.Variables[key]
	createdAt := current.CreatedAt
	if createdAt == "" {
		createdAt = utcNow()
	}
	environment.Variables[key] = VariableEntry{
		Value:     storedValue,
		IsSecret:  encrypt,
		CreatedAt: createdAt,
		UpdatedAt: utcNow(),
	}
	cfg.Environments[resolved] = environment

	if err := s.store.Save(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (s *EnvxService) ListVariables(envName string, showSecrets bool) ([]VariableRow, error) {
	cfg, err := s.store.Load()
	if err != nil {
		return nil, err
	}
	resolved, err := resolveEnvironment(cfg, envName)
	if err != nil {
		return nil, err
	}

	environment := cfg.Environments[resolved]
	keys := make([]string, 0, len(environment.Variables))
	for key := range environment.Variables {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	rows := make([]VariableRow, 0, len(keys))
	for _, key := range keys {
		entry := environment.Variables[key]
		display := ""
		if showSecrets {
			display, err = s.crypto.Decrypt(entry.Value)
			if err != nil {
				return nil, err
			}
		} else {
			display = crypto.RedactValue(entry.Value, entry.IsSecret)
		}
		rows = append(rows, VariableRow{
			Key:         key,
			Value:       display,
			Secret:      strconv.FormatBool(entry.IsSecret),
			Environment: resolved,
		})
	}
	return rows, nil
}

func (s *EnvxService) CreateEnvironment(envName string) error {
	if err := ValidateEnvName(envName); err != nil {
		return err
	}
	cfg, err := s.store.Load()
	if err != nil {
		return err
	}
	if _, ok := cfg.Environments[envName]; ok {
		return fmt.Errorf("%w: environment '%s' already exists", errs.ErrEnvironmentConflict, envName)
	}
	cfg.Environments[envName] = EnvironmentSpec{Name: envName, Variables: map[string]VariableEntry{}}
	return s.store.Save(cfg)
}

func (s *EnvxService) UseEnvironment(envName string) error {
	cfg, err := s.store.Load()
	if err != nil {
		return err
	}
	if _, err := resolveEnvironment(cfg, envName); err != nil {
		return err
	}
	cfg.ActiveEnv = envName
	return s.store.Save(cfg)
}

func (s *EnvxService) DeleteEnvironment(envName string) error {
	cfg, err := s.store.Load()
	if err != nil {
		return err
	}
	if _, err := resolveEnvironment(cfg, envName); err != nil {
		return err
	}
	if envName == cfg.ActiveEnv {
		return fmt.Errorf("%w: cannot delete the active environment", errs.ErrEnvironmentConflict)
	}
	delete(cfg.Environments, envName)
	return s.store.Save(cfg)
}

func (s *EnvxService) ListEnvironments() ([]EnvironmentRow, error) {
	cfg, err := s.store.Load()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(cfg.Environments))
	for name := range cfg.Environments {
		names = append(names, name)
	}
	sort.Strings(names)

	rows := make([]EnvironmentRow, 0, len(names))
	for _, name := range names {
		rows = append(rows, EnvironmentRow{
			Name:      name,
			Active:    strconv.FormatBool(name == cfg.ActiveEnv),
			Variables: strconv.Itoa(len(cfg.Environments[name].Variables)),
		})
	}
	return rows, nil
}

func (s *EnvxService) ResolveRuntimeEnv(envName string) (string, map[string]string, error) {
	cfg, err := s.store.Load()
	if err != nil {
		return "", nil, err
	}
	resolved, err := resolveEnvironment(cfg, envName)
	if err != nil {
		return "", nil, err
	}

	envVars := make(map[string]string, len(cfg.Environments[resolved].Variables))
	for key, entry := range cfg.Environments[resolved].Variables {
		value, err := s.crypto.Decrypt(entry.Value)
		if err != nil {
			return "", nil, err
		}
		envVars[key] = value
	}
	return resolved, envVars, nil
}

func (s *EnvxService) RunCommand(argv []string, shellCmd, envName string) (int, error) {
	_, envVars, err := s.ResolveRuntimeEnv(envName)
	if err != nil {
		return 1, err
	}
	return s.runner.Run(argv, shellCmd, envVars)
}

func resolveEnvironment(cfg *ConfigFile, requested string) (string, error) {
	name := requested
	if name == "" {
		name = cfg.ActiveEnv
	}
	if _, ok := cfg.Environments[name]; !ok {
		return "", fmt.Errorf("%w: environment '%s' was not found", errs.ErrEnvironmentNotFound, name)
	}
	return name, nil
}
