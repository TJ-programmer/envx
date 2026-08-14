package web

import (
	"embed"
	"encoding/json"
	"net/http"

	"envx/internal/core"
)

//go:embed assets/index.html
var assets embed.FS

type Server struct {
	service *core.EnvxService
}

func New(service *core.EnvxService) http.Handler {
	s := &Server{service: service}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/status", s.handleStatus)
	mux.HandleFunc("POST /api/init", s.handleInit)
	mux.HandleFunc("GET /api/environments", s.handleEnvironments)
	mux.HandleFunc("POST /api/environments", s.handleCreateEnvironment)
	mux.HandleFunc("POST /api/environments/use", s.handleUseEnvironment)
	mux.HandleFunc("DELETE /api/environments/{name}", s.handleDeleteEnvironment)
	mux.HandleFunc("GET /api/variables", s.handleVariables)
	mux.HandleFunc("POST /api/variables", s.handleSetVariable)
	mux.HandleFunc("DELETE /api/variables/{key}", s.handleUnsetVariable)
	mux.HandleFunc("GET /api/export", s.handleExport)
	mux.HandleFunc("/", s.handleIndex)
	return mux
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	data, err := assets.ReadFile("assets/index.html")
	if err != nil {
		http.Error(w, "embedded UI missing", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func writeErr(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func envNames(cfg *core.ConfigFile) []string {
	names := make([]string, 0, len(cfg.Environments))
	for name := range cfg.Environments {
		names = append(names, name)
	}
	return names
}
