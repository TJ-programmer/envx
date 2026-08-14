package core_test

import (
	"errors"
	"testing"

	"envx/internal/bootstrap"
	"envx/internal/core"
	"envx/internal/crypto"
	"envx/internal/errs"
)

func serviceFor(t *testing.T) *core.EnvxService {
	return bootstrap.BuildService(t.TempDir())
}

func TestSecretValuesAreRedactedByDefault(t *testing.T) {
	service := serviceFor(t)
	if _, err := service.InitProject("dev", false); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SetVariable("API_KEY", "super-secret", "", true, false); err != nil {
		t.Fatal(err)
	}

	rows, err := service.ListVariables("", false)
	if err != nil {
		t.Fatal(err)
	}
	if rows[0].Value != crypto.RedactedValue {
		t.Errorf("value = %q, want %q", rows[0].Value, crypto.RedactedValue)
	}
}

func TestShowSecretsRevealsPlaintext(t *testing.T) {
	service := serviceFor(t)
	if _, err := service.InitProject("dev", false); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SetVariable("API_KEY", "super-secret", "", true, false); err != nil {
		t.Fatal(err)
	}

	rows, err := service.ListVariables("", true)
	if err != nil {
		t.Fatal(err)
	}
	if rows[0].Value != "super-secret" {
		t.Errorf("value = %q, want %q", rows[0].Value, "super-secret")
	}
}

func TestEnvironmentLifecycle(t *testing.T) {
	service := serviceFor(t)
	if _, err := service.InitProject("dev", false); err != nil {
		t.Fatal(err)
	}
	if err := service.CreateEnvironment("prod"); err != nil {
		t.Fatal(err)
	}
	if err := service.UseEnvironment("prod"); err != nil {
		t.Fatal(err)
	}

	rows, err := service.ListEnvironments()
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if row.Name == "prod" && row.Active != "true" {
			t.Errorf("prod should be active: %+v", row)
		}
	}
}

func TestDeleteActiveEnvironmentIsRejected(t *testing.T) {
	service := serviceFor(t)
	if _, err := service.InitProject("dev", false); err != nil {
		t.Fatal(err)
	}

	err := service.DeleteEnvironment("dev")
	if !errors.Is(err, errs.ErrEnvironmentConflict) {
		t.Fatalf("err = %v", err)
	}
}

func TestMissingEnvironmentRaises(t *testing.T) {
	service := serviceFor(t)
	if _, err := service.InitProject("dev", false); err != nil {
		t.Fatal(err)
	}

	err := service.UseEnvironment("prod")
	if !errors.Is(err, errs.ErrEnvironmentNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestPlainVariableNotEncrypted(t *testing.T) {
	service := serviceFor(t)
	if _, err := service.InitProject("dev", false); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SetVariable("PORT", "8000", "", false, false); err != nil {
		t.Fatal(err)
	}

	cfg, err := service.Load()
	if err != nil {
		t.Fatal(err)
	}
	entry := cfg.Environments["dev"].Variables["PORT"]
	if entry.IsSecret {
		t.Error("plain variable should not be marked secret")
	}
	if entry.Value != "8000" {
		t.Errorf("value = %q", entry.Value)
	}
}
