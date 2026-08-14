package store

import (
	"fmt"
	"strings"
	"time"

	"envx/internal/errs"
)

func migrateConfig(raw map[string]any) (map[string]any, error) {
	if version, ok := raw["version"]; ok {
		ver, _ := version.(float64)
		if ver > 2 {
			return nil, fmt.Errorf("%w: unsupported config schema version %v", errs.ErrConfigVersion, version)
		}
		if _, ok := raw["encryption"]; !ok {
			raw["encryption"] = map[string]any{"default_encrypt": false, "key_backend": "file"}
		}
		raw["version"] = 2.0
		return raw, nil
	}

	activeEnv := "dev"
	if value, ok := raw["active_env"].(string); ok {
		activeEnv = value
	}

	environments := map[string]any{}
	if rawEnvs, ok := raw["environments"].(map[string]any); ok {
		for envName, rawEnv := range rawEnvs {
			envMap, ok := rawEnv.(map[string]any)
			if !ok {
				continue
			}
			variables := map[string]any{}
			if rawVars, has := envMap["variables"].(map[string]any); has {
				for key, value := range rawVars {
					if entry, ok := value.(map[string]any); ok {
						variables[key] = entry
					}
				}
			} else {
				for key, value := range envMap {
					if str, ok := value.(string); ok {
						variables[key] = legacyVariableEntry(str)
					}
				}
			}
			environments[envName] = map[string]any{
				"name":      envName,
				"variables": variables,
			}
		}
	}

	return map[string]any{
		"version":      2.0,
		"active_env":   activeEnv,
		"encryption":   map[string]any{"default_encrypt": false, "key_backend": "file"},
		"environments": environments,
		"metadata":     map[string]any{"migrated_from_legacy": true},
	}, nil
}

func legacyVariableEntry(value string) map[string]any {
	isSecret := strings.HasPrefix(value, "enc:")
	now := utcNow()
	return map[string]any{
		"value":      value,
		"is_secret":  isSecret,
		"created_at": now,
		"updated_at": now,
	}
}

func utcNow() string {
	return time.Now().UTC().Truncate(time.Second).Format(time.RFC3339)
}
