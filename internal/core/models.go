package core

import "time"

const SchemaVersion = 2

func utcNow() string {
	return time.Now().UTC().Truncate(time.Second).Format(time.RFC3339)
}

type VariableEntry struct {
	Value     string `json:"value"`
	IsSecret  bool   `json:"is_secret"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type EnvironmentSpec struct {
	Name      string                   `json:"name"`
	Variables map[string]VariableEntry `json:"variables"`
}

type EncryptionSettings struct {
	DefaultEncrypt bool   `json:"default_encrypt"`
	KeyBackend     string `json:"key_backend"`
}

type MigrationSettings struct {
	OverlayDotenv bool `json:"overlay_dotenv"`
}

type ConfigFile struct {
	Version      int                        `json:"version"`
	ActiveEnv    string                     `json:"active_env"`
	Encryption   EncryptionSettings         `json:"encryption"`
	Migration    MigrationSettings          `json:"migration"`
	Environments map[string]EnvironmentSpec `json:"environments"`
	Metadata     map[string]any             `json:"metadata"`
}

func NewConfigFile(envName string) *ConfigFile {
	return &ConfigFile{
		Version:   SchemaVersion,
		ActiveEnv: envName,
		Encryption: EncryptionSettings{
			DefaultEncrypt: false,
			KeyBackend:     "file",
		},
		Migration: MigrationSettings{
			OverlayDotenv: false,
		},
		Environments: map[string]EnvironmentSpec{
			envName: {Name: envName, Variables: map[string]VariableEntry{}},
		},
		Metadata: map[string]any{"created_at": utcNow()},
	}
}
