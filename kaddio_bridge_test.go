package kaddio_bridge_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"kaddio-bridge/internal/api"
	"kaddio-bridge/internal/auth"
	"kaddio-bridge/internal/config"
	"kaddio-bridge/internal/process"
)

func TestTokenGeneration(t *testing.T) {
	token, err := auth.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(token) != 32 {
		t.Errorf("token length = %d, want 32 (16 bytes hex)", len(token))
	}

	for _, c := range token {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("token contains non-hex character: %c", c)
			break
		}
	}

	token2, _ := auth.Generate()
	if token == token2 {
		t.Error("two generated tokens should not be equal")
	}
}

func TestTokenValidation(t *testing.T) {
	token, _ := auth.Generate()

	if !auth.Validate(token, token) {
		t.Error("Validate() should return true for matching tokens")
	}

	if auth.Validate(token, "wrong-token") {
		t.Error("Validate() should return false for non-matching tokens")
	}

	if auth.Validate("", token) {
		t.Error("Validate() should return false for empty token")
	}
}

func TestConfigCreation(t *testing.T) {
	dir := t.TempDir()

	os.Setenv("KADDIO_BRIDGE_CONFIG_DIR", dir)
	defer os.Unsetenv("KADDIO_BRIDGE_CONFIG_DIR")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Address != "127.0.0.1:38471" {
		t.Errorf("Address = %q, want %q", cfg.Address, "127.0.0.1:38471")
	}

	if cfg.Token == "" {
		t.Error("Token should not be empty")
	}

	if len(cfg.Token) != 32 {
		t.Errorf("Token length = %d, want 32", len(cfg.Token))
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("config dir stat error: %v", err)
	}

	if info.Mode().Perm() != 0700 {
		t.Errorf("config dir permissions = %v, want 0700", info.Mode().Perm())
	}

	cfgFile := filepath.Join(dir, "config.json")
	finfo, err := os.Stat(cfgFile)
	if err != nil {
		t.Fatalf("config file stat error: %v", err)
	}

	if finfo.Mode().Perm() != 0600 {
		t.Errorf("config file permissions = %v, want 0600", finfo.Mode().Perm())
	}
}

func TestConfigLoadExisting(t *testing.T) {
	dir := t.TempDir()

	os.Setenv("KADDIO_BRIDGE_CONFIG_DIR", dir)
	defer os.Unsetenv("KADDIO_BRIDGE_CONFIG_DIR")

	cfg1, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	cfg2, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg1.Token != cfg2.Token {
		t.Error("loading config twice should return the same token")
	}

	if cfg1.Address != cfg2.Address {
		t.Error("loading config twice should return the same address")
	}
}

func TestStatusEndpoint(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("KADDIO_BRIDGE_CONFIG_DIR", dir)
	defer os.Unsetenv("KADDIO_BRIDGE_CONFIG_DIR")

	cfg, _ := config.Load()
	srv := api.New(cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/status", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("status = %q, want %q", resp["status"], "ok")
	}

	if resp["version"] == "" {
		t.Error("version should not be empty")
	}
}

func TestPingEndpointUnauthorized(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("KADDIO_BRIDGE_CONFIG_DIR", dir)
	defer os.Unsetenv("KADDIO_BRIDGE_CONFIG_DIR")

	cfg, _ := config.Load()
	srv := api.New(cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestPingEndpointAuthorized(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("KADDIO_BRIDGE_CONFIG_DIR", dir)
	defer os.Unsetenv("KADDIO_BRIDGE_CONFIG_DIR")

	cfg, _ := config.Load()
	srv := api.New(cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if resp["ok"] != true {
		t.Errorf("ok = %v, want true", resp["ok"])
	}

	if resp["version"] == "" {
		t.Error("version should not be empty")
	}
}

func TestProcessEndpointUnauthorized(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("KADDIO_BRIDGE_CONFIG_DIR", dir)
	defer os.Unsetenv("KADDIO_BRIDGE_CONFIG_DIR")

	cfg, _ := config.Load()
	srv := api.New(cfg)

	body := `{"path":"/bin/echo","args":["hello"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/process", strings.NewReader(body))
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestProcessEndpointBadToken(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("KADDIO_BRIDGE_CONFIG_DIR", dir)
	defer os.Unsetenv("KADDIO_BRIDGE_CONFIG_DIR")

	cfg, _ := config.Load()
	srv := api.New(cfg)

	body := `{"path":"/bin/echo","args":["hello"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/process", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer bad-token")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestProcessEndpointBadJSON(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("KADDIO_BRIDGE_CONFIG_DIR", dir)
	defer os.Unsetenv("KADDIO_BRIDGE_CONFIG_DIR")

	cfg, _ := config.Load()
	srv := api.New(cfg)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/process", strings.NewReader("not json"))
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["started"] != false {
		t.Errorf("started = %v, want false", resp["started"])
	}
}

func TestProcessEndpointInvalidPath(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("KADDIO_BRIDGE_CONFIG_DIR", dir)
	defer os.Unsetenv("KADDIO_BRIDGE_CONFIG_DIR")

	cfg, _ := config.Load()
	srv := api.New(cfg)

	body := `{"path":"/nonexistent/app","args":[]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/process", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["started"] != false {
		t.Errorf("started = %v, want false", resp["started"])
	}
}

func TestProcessRequestParsing(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
		wantPath string
	}{
		{
			name:     "minimal",
			json:     `{"path":"/bin/echo"}`,
			wantErr:  false,
			wantPath: "/bin/echo",
		},
		{
			name:    "empty path",
			json:    `{"path":""}`,
			wantErr: false,
		},
		{
			name:     "with args",
			json:     `{"path":"/bin/echo","args":["-n","hello"]}`,
			wantErr:  false,
			wantPath: "/bin/echo",
		},
		{
			name:     "with cwd",
			json:     `{"path":"/bin/echo","cwd":"/tmp"}`,
			wantErr:  false,
			wantPath: "/bin/echo",
		},
		{
			name:    "invalid json",
			json:    `{bad`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req process.Request
			err := json.Unmarshal([]byte(tt.json), &req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && req.Path != tt.wantPath {
				t.Errorf("Path = %q, want %q", req.Path, tt.wantPath)
			}
		})
	}
}
