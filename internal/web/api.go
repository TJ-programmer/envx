package web

import (
	"encoding/json"
	"net/http"

	"envx/internal/buildinfo"
	"envx/internal/gitignore"
)

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if !s.service.IsInitialized() {
		writeJSON(w, http.StatusOK, map[string]any{"initialized": false, "version": buildinfo.Version})
		return
	}
	cfg, err := s.service.Load()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"initialized":     true,
		"version":         buildinfo.Version,
		"root":            s.service.StoreRoot(),
		"active_env":      cfg.ActiveEnv,
		"environments":    envNames(cfg),
		"default_encrypt": cfg.Encryption.DefaultEncrypt,
		"key_backend":     cfg.Encryption.KeyBackend,
		"overlay_dotenv":  cfg.Migration.OverlayDotenv,
		"schema_version":  cfg.Version,
	})
}

func (s *Server) handleInit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		EnvName string `json:"env_name"`
		Force   bool   `json:"force"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	if req.EnvName == "" {
		req.EnvName = "dev"
	}
	if _, err := s.service.InitProject(req.EnvName, req.Force); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	if err := gitignore.Ensure(s.service.StoreRoot(), false); err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"ok": true, "active_env": req.EnvName})
}

func (s *Server) handleEnvironments(w http.ResponseWriter, r *http.Request) {
	rows, err := s.service.ListEnvironments()
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"environments": rows})
}

func (s *Server) handleCreateEnvironment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	if err := s.service.CreateEnvironment(req.Name); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"ok": true, "name": req.Name})
}

func (s *Server) handleUseEnvironment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	if err := s.service.UseEnvironment(req.Name); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "name": req.Name})
}

func (s *Server) handleDeleteEnvironment(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := s.service.DeleteEnvironment(name); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "name": name})
}

func (s *Server) handleVariables(w http.ResponseWriter, r *http.Request) {
	envName := r.URL.Query().Get("env")
	showSecrets := r.URL.Query().Get("show_secrets") == "true"
	rows, err := s.service.ListVariables(envName, showSecrets)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"environment": envName, "variables": rows})
}

func (s *Server) handleSetVariable(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key    string `json:"key"`
		Value  string `json:"value"`
		Env    string `json:"env"`
		Secret bool   `json:"secret"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	if _, err := s.service.SetVariable(req.Key, req.Value, req.Env, req.Secret, !req.Secret); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"ok": true, "key": req.Key, "secret": req.Secret})
}

func (s *Server) handleUnsetVariable(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	envName := r.URL.Query().Get("env")
	if err := s.service.UnsetVariable(key, envName); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "key": key})
}

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	envName := r.URL.Query().Get("env")
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "dotenv"
	}
	out, err := s.service.ExportEnv(envName, format)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(out))
}
