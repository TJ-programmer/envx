package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"envx/internal/bootstrap"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	service := bootstrap.BuildService(t.TempDir())
	if _, err := service.InitProject("dev", false); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SetVariable("PORT", "8000", "", false, false); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SetVariable("DB_PASSWORD", "hunter2", "", true, false); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(New(service))
	t.Cleanup(srv.Close)
	return srv
}

func doJSON(t *testing.T, method, url string, body any) (int, map[string]any) {
	t.Helper()
	var reader *bytes.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		reader = bytes.NewReader(data)
	} else {
		reader = bytes.NewReader(nil)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var payload map[string]any
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	return res.StatusCode, payload
}

func TestIndexServesHTML(t *testing.T) {
	srv := newTestServer(t)
	res, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", res.StatusCode)
	}
	if ct := res.Header.Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Fatalf("content-type = %q", ct)
	}
}

func TestStatusShowsActiveEnv(t *testing.T) {
	srv := newTestServer(t)
	code, payload := doJSON(t, "GET", srv.URL+"/api/status", nil)
	if code != http.StatusOK {
		t.Fatalf("status code = %d", code)
	}
	if payload["initialized"] != true {
		t.Fatalf("initialized = %v", payload["initialized"])
	}
	if payload["active_env"] != "dev" {
		t.Fatalf("active_env = %v", payload["active_env"])
	}
}

func TestVariablesRedactedByDefault(t *testing.T) {
	srv := newTestServer(t)
	code, payload := doJSON(t, "GET", srv.URL+"/api/variables?env=dev", nil)
	if code != http.StatusOK {
		t.Fatalf("status code = %d", code)
	}
	raw := payload["variables"].([]any)
	byKey := map[string]map[string]any{}
	for _, item := range raw {
		row := item.(map[string]any)
		byKey[row["key"].(string)] = row
	}
	if byKey["DB_PASSWORD"]["value"] != "********" {
		t.Fatalf("secret should be redacted: %v", byKey["DB_PASSWORD"]["value"])
	}
	if byKey["PORT"]["value"] != "8000" {
		t.Fatalf("plain value = %v", byKey["PORT"]["value"])
	}

	code, payload = doJSON(t, "GET", srv.URL+"/api/variables?env=dev&show_secrets=true", nil)
	if code != http.StatusOK {
		t.Fatalf("status code = %d", code)
	}
	raw = payload["variables"].([]any)
	for _, item := range raw {
		row := item.(map[string]any)
		if row["key"] == "DB_PASSWORD" && row["value"] != "hunter2" {
			t.Fatalf("secret should be revealed with show_secrets: %v", row["value"])
		}
	}
}

func TestSetAndUnsetVariableViaAPI(t *testing.T) {
	srv := newTestServer(t)

	code, _ := doJSON(t, "POST", srv.URL+"/api/variables", map[string]any{
		"key": "FEATURE", "value": "on", "env": "dev", "secret": false,
	})
	if code != http.StatusCreated {
		t.Fatalf("set status = %d", code)
	}

	code, payload := doJSON(t, "GET", srv.URL+"/api/variables?env=dev", nil)
	if code != http.StatusOK {
		t.Fatalf("list status = %d", code)
	}
	found := false
	for _, item := range payload["variables"].([]any) {
		if item.(map[string]any)["key"] == "FEATURE" {
			found = true
		}
	}
	if !found {
		t.Fatal("FEATURE not found after set")
	}

	code, _ = doJSON(t, "DELETE", srv.URL+"/api/variables/FEATURE?env=dev", nil)
	if code != http.StatusOK {
		t.Fatalf("unset status = %d", code)
	}

	code, payload = doJSON(t, "GET", srv.URL+"/api/variables?env=dev", nil)
	if code != http.StatusOK {
		t.Fatalf("list status = %d", code)
	}
	for _, item := range payload["variables"].([]any) {
		if item.(map[string]any)["key"] == "FEATURE" {
			t.Fatal("FEATURE should be gone after unset")
		}
	}
}

func TestEnvironmentLifecycleViaAPI(t *testing.T) {
	srv := newTestServer(t)

	code, _ := doJSON(t, "POST", srv.URL+"/api/environments", map[string]any{"name": "prod"})
	if code != http.StatusCreated {
		t.Fatalf("create status = %d", code)
	}

	code, _ = doJSON(t, "POST", srv.URL+"/api/environments/use", map[string]any{"name": "prod"})
	if code != http.StatusOK {
		t.Fatalf("use status = %d", code)
	}

	code, payload := doJSON(t, "GET", srv.URL+"/api/status", nil)
	if code != http.StatusOK {
		t.Fatalf("status code = %d", code)
	}
	if payload["active_env"] != "prod" {
		t.Fatalf("active_env = %v", payload["active_env"])
	}

	code, _ = doJSON(t, "DELETE", srv.URL+"/api/environments/prod", nil)
	if code != http.StatusBadRequest {
		t.Fatalf("deleting the active environment should be rejected, got %d", code)
	}

	code, _ = doJSON(t, "POST", srv.URL+"/api/environments/use", map[string]any{"name": "dev"})
	if code != http.StatusOK {
		t.Fatalf("use status = %d", code)
	}
	code, _ = doJSON(t, "DELETE", srv.URL+"/api/environments/prod", nil)
	if code != http.StatusOK {
		t.Fatalf("delete non-active status = %d", code)
	}

	code, payload = doJSON(t, "GET", srv.URL+"/api/status", nil)
	if code != http.StatusOK {
		t.Fatalf("status code = %d", code)
	}
	envs := payload["environments"].([]any)
	if len(envs) != 1 {
		t.Fatalf("environments after delete = %v", payload["environments"])
	}
}

func TestExportEndpoint(t *testing.T) {
	srv := newTestServer(t)
	res, err := http.Get(srv.URL + "/api/export?env=dev&format=dotenv")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", res.StatusCode)
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Body)
	if !strings.Contains(buf.String(), "PORT=8000") {
		t.Fatalf("export output missing PORT:\n%s", buf.String())
	}
}

func TestErrorsReturnedAsJSON(t *testing.T) {
	srv := newTestServer(t)
	code, payload := doJSON(t, "POST", srv.URL+"/api/variables", map[string]any{
		"key": "BAD KEY", "value": "x", "env": "dev", "secret": false,
	})
	if code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", code)
	}
	if payload["error"] == nil {
		t.Fatal("error body missing")
	}
}

func TestNotInitializedStatus(t *testing.T) {
	service := bootstrap.BuildService(t.TempDir())
	srv := httptest.NewServer(New(service))
	defer srv.Close()

	res, err := http.Get(srv.URL + "/api/status")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var payload map[string]any
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if payload["initialized"] != false {
		t.Fatalf("initialized = %v", payload["initialized"])
	}
}
