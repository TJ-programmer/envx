package core

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"envx/internal/crypto"
	"envx/internal/errs"
)

type ConfigStore interface {
	Exists() bool
	Initialize(cfg *ConfigFile, force bool) error
	Load() (*ConfigFile, error)
	Save(cfg *ConfigFile) error
	Root() string
}

type Crypter interface {
	EnsureKey() error
	Encrypt(value string) (string, error)
	Decrypt(value string) (string, error)
	WriteKey(raw string) error
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

type DiffRow struct {
	Key    string
	ValueA string
	ValueB string
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

func (s *EnvxService) RunCommand(argv []string, shellCmd, envName string, overlay bool) (int, int, error) {
	_, envVars, err := s.ResolveRuntimeEnv(envName)
	if err != nil {
		return 1, 0, err
	}
	overlayCount, err := s.applyDotenvOverlay(envVars, overlay)
	if err != nil {
		return 1, 0, err
	}
	code, err := s.runner.Run(argv, shellCmd, envVars)
	if err != nil {
		return 1, overlayCount, err
	}
	return code, overlayCount, nil
}

func (s *EnvxService) applyDotenvOverlay(envVars map[string]string, overlay bool) (int, error) {
	if !overlay {
		cfg, err := s.store.Load()
		if err != nil {
			return 0, err
		}
		if !cfg.Migration.OverlayDotenv {
			return 0, nil
		}
	}
	dotenvPath := filepath.Join(s.store.Root(), ".env")
	data, err := os.ReadFile(dotenvPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	entries, err := parseDotenv(string(data))
	if err != nil {
		return 0, fmt.Errorf("legacy .env overlay: %w", err)
	}
	count := 0
	for _, kv := range entries {
		if _, exists := envVars[kv.key]; !exists {
			envVars[kv.key] = kv.value
			count++
		}
	}
	return count, nil
}

func (s *EnvxService) RotateKey() (int, error) {
	cfg, err := s.store.Load()
	if err != nil {
		return 0, err
	}

	type loc struct{ env, key string }
	plaintexts := make(map[loc]string)
	for envName, environment := range cfg.Environments {
		for key, entry := range environment.Variables {
			if !entry.IsSecret {
				continue
			}
			plain, err := s.crypto.Decrypt(entry.Value)
			if err != nil {
				return 0, err
			}
			plaintexts[loc{env: envName, key: key}] = plain
		}
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return 0, err
	}
	encoded := base64.URLEncoding.EncodeToString(raw)
	if err := s.crypto.WriteKey(encoded); err != nil {
		return 0, err
	}

	for envName, environment := range cfg.Environments {
		for key, entry := range environment.Variables {
			if !entry.IsSecret {
				continue
			}
			newValue, err := s.crypto.Encrypt(plaintexts[loc{env: envName, key: key}])
			if err != nil {
				return 0, err
			}
			entry.Value = newValue
			environment.Variables[key] = entry
		}
		cfg.Environments[envName] = environment
	}

	if err := s.store.Save(cfg); err != nil {
		return 0, err
	}
	return len(plaintexts), nil
}

func (s *EnvxService) GetVariable(key, envName string, showSecret bool) (string, error) {
	if err := ValidateVariableName(key); err != nil {
		return "", err
	}
	cfg, err := s.store.Load()
	if err != nil {
		return "", err
	}
	resolved, err := resolveEnvironment(cfg, envName)
	if err != nil {
		return "", err
	}
	entry, ok := cfg.Environments[resolved].Variables[key]
	if !ok {
		return "", fmt.Errorf("%w: variable '%s' was not found in environment '%s'", errs.ErrVariableNotFound, key, resolved)
	}
	if entry.IsSecret && !showSecret {
		return crypto.RedactedValue, nil
	}
	return s.crypto.Decrypt(entry.Value)
}

func (s *EnvxService) UnsetVariable(key, envName string) error {
	if err := ValidateVariableName(key); err != nil {
		return err
	}
	cfg, err := s.store.Load()
	if err != nil {
		return err
	}
	resolved, err := resolveEnvironment(cfg, envName)
	if err != nil {
		return err
	}
	environment := cfg.Environments[resolved]
	if _, ok := environment.Variables[key]; !ok {
		return fmt.Errorf("%w: variable '%s' was not found in environment '%s'", errs.ErrVariableNotFound, key, resolved)
	}
	delete(environment.Variables, key)
	cfg.Environments[resolved] = environment
	return s.store.Save(cfg)
}

func (s *EnvxService) ImportEnvFile(path, envName string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	cfg, err := s.store.Load()
	if err != nil {
		return 0, err
	}
	resolved, err := resolveEnvironment(cfg, envName)
	if err != nil {
		return 0, err
	}

	entries, err := parseDotenv(string(data))
	if err != nil {
		return 0, err
	}
	for _, kv := range entries {
		existing := cfg.Environments[resolved].Variables[kv.key]
		secret := existing.IsSecret || IsSensitiveName(kv.key)
		if _, err := s.SetVariable(kv.key, kv.value, resolved, secret, false); err != nil {
			return 0, err
		}
	}
	return len(entries), nil
}

func (s *EnvxService) ExportEnv(envName, format string) (string, error) {
	_, envVars, err := s.ResolveRuntimeEnv(envName)
	if err != nil {
		return "", err
	}
	keys := make([]string, 0, len(envVars))
	for key := range envVars {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	switch format {
	case "shell":
		var b strings.Builder
		for _, key := range keys {
			fmt.Fprintf(&b, "export %s=%s\n", key, shellQuote(envVars[key]))
		}
		return b.String(), nil
	case "dotenv":
		var b strings.Builder
		for _, key := range keys {
			fmt.Fprintf(&b, "%s=%s\n", key, dotenvQuote(envVars[key]))
		}
		return b.String(), nil
	case "json":
		data, err := json.MarshalIndent(envVars, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data) + "\n", nil
	default:
		return "", fmt.Errorf("unknown export format '%s' (use shell, dotenv, or json)", format)
	}
}

func (s *EnvxService) DiffEnvironments(envA, envB string) ([]DiffRow, error) {
	cfg, err := s.store.Load()
	if err != nil {
		return nil, err
	}
	a, err := resolveEnvironment(cfg, envA)
	if err != nil {
		return nil, err
	}
	b, err := resolveEnvironment(cfg, envB)
	if err != nil {
		return nil, err
	}
	va, err := s.decryptedMap(cfg, a)
	if err != nil {
		return nil, err
	}
	vb, err := s.decryptedMap(cfg, b)
	if err != nil {
		return nil, err
	}

	keys := make(map[string]struct{}, len(va)+len(vb))
	for key := range va {
		keys[key] = struct{}{}
	}
	for key := range vb {
		keys[key] = struct{}{}
	}
	sorted := make([]string, 0, len(keys))
	for key := range keys {
		sorted = append(sorted, key)
	}
	sort.Strings(sorted)

	var rows []DiffRow
	for _, key := range sorted {
		valueA, inA := va[key]
		valueB, inB := vb[key]
		if inA == inB && (!inA || valueA == valueB) {
			continue
		}
		rowA, rowB := "-", "-"
		if inA {
			rowA = valueA
		}
		if inB {
			rowB = valueB
		}
		rows = append(rows, DiffRow{Key: key, ValueA: rowA, ValueB: rowB})
	}
	return rows, nil
}

func (s *EnvxService) GetSetting(key string) (string, error) {
	cfg, err := s.store.Load()
	if err != nil {
		return "", err
	}
	switch key {
	case "encryption.default_encrypt":
		return strconv.FormatBool(cfg.Encryption.DefaultEncrypt), nil
	case "key_backend":
		return cfg.Encryption.KeyBackend, nil
	case "migration.overlay_dotenv":
		return strconv.FormatBool(cfg.Migration.OverlayDotenv), nil
	case "active_env":
		return cfg.ActiveEnv, nil
	default:
		return "", fmt.Errorf("unknown setting '%s'", key)
	}
}

func (s *EnvxService) SetDefaultEncrypt(value bool) error {
	cfg, err := s.store.Load()
	if err != nil {
		return err
	}
	cfg.Encryption.DefaultEncrypt = value
	return s.store.Save(cfg)
}

func (s *EnvxService) SetOverlayDotenv(value bool) error {
	cfg, err := s.store.Load()
	if err != nil {
		return err
	}
	cfg.Migration.OverlayDotenv = value
	return s.store.Save(cfg)
}

func (s *EnvxService) decryptedMap(cfg *ConfigFile, envName string) (map[string]string, error) {
	out := make(map[string]string, len(cfg.Environments[envName].Variables))
	for key, entry := range cfg.Environments[envName].Variables {
		value, err := s.crypto.Decrypt(entry.Value)
		if err != nil {
			return nil, err
		}
		out[key] = value
	}
	return out, nil
}

type keyValue struct {
	key   string
	value string
}

func parseDotenv(content string) ([]keyValue, error) {
	var out []keyValue
	for i, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		idx := strings.Index(line, "=")
		if idx < 0 {
			return nil, fmt.Errorf("invalid dotenv line %d: missing '='", i+1)
		}
		key := strings.TrimSpace(line[:idx])
		if err := ValidateVariableName(key); err != nil {
			return nil, fmt.Errorf("line %d: %w", i+1, err)
		}
		value := strings.TrimSpace(line[idx+1:])
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}
		out = append(out, keyValue{key: key, value: value})
	}
	return out, nil
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}

func dotenvQuote(value string) string {
	if value == "" {
		return `""`
	}
	if !strings.ContainsAny(value, " \t\n#\"'\\") {
		return value
	}
	quoted := strings.ReplaceAll(value, `\`, `\\`)
	quoted = strings.ReplaceAll(quoted, `"`, `\"`)
	return `"` + quoted + `"`
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
