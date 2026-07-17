package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"locallaunch/internal/auth"
	"locallaunch/internal/config"
	"locallaunch/internal/process"
)

const version = "0.1.1"

type Server struct {
	cfg    *config.Config
	mux    *http.ServeMux
	server *http.Server
}

func New(cfg *config.Config) *Server {
	s := &Server{
		cfg: cfg,
		mux: http.NewServeMux(),
	}

	s.mux.HandleFunc("/api/v1/status", s.handleStatus)
	s.mux.HandleFunc("/api/v1/process", s.handleProcess)

	s.server = &http.Server{
		Addr:    cfg.Address,
		Handler: s.mux,
	}

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", s.cfg.Address)
	if err != nil {
		return fmt.Errorf("listening on %s: %w", s.cfg.Address, err)
	}

	log.Printf("Listening:\n%s\n", s.cfg.Address)

	return s.server.Serve(ln)
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"version": version,
	})
}

func (s *Server) handleProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !s.authenticate(r) {
		http.Error(w, `{"started":false,"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req process.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"started": false,
			"error":   "invalid request body: " + err.Error(),
		})
		return
	}

	log.Printf("Launching:\n%s\n\nArguments:\n%v\n", req.Path, req.Args)

	pid, err := process.Launch(&req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"started": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"started": true,
		"pid":     pid,
	})
}

func (s *Server) authenticate(r *http.Request) bool {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return false
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		return false
	}

	token := strings.TrimPrefix(authHeader, prefix)
	return auth.Validate(token, s.cfg.Token)
}
