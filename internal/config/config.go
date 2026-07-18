package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"kaddio-bridge/internal/auth"
)

const (
	dirPermissions  = 0700
	filePermissions = 0600
	appName         = "kaddio-bridge"
	configFile      = "config.json"
	defaultAddress  = "127.0.0.1:38471"
)

type Config struct {
	Address string `json:"address"`
	Token   string `json:"token"`
}

func Dir() (string, error) {
	if override := os.Getenv("KADDIO_BRIDGE_CONFIG_DIR"); override != "" {
		return override, nil
	}

	var base string

	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("getting home directory: %w", err)
		}
		base = filepath.Join(home, "Library", "Application Support")
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("getting home directory: %w", err)
			}
			base = filepath.Join(home, "AppData", "Roaming")
		} else {
			base = appData
		}
	default:
		xdg := os.Getenv("XDG_CONFIG_HOME")
		if xdg != "" {
			base = xdg
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("getting home directory: %w", err)
			}
			base = filepath.Join(home, ".config")
		}
	}

	return filepath.Join(base, appName), nil
}

func path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFile), nil
}

func Load() (*Config, error) {
	cfgPath, err := path()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(cfgPath)
	if err == nil {
		var cfg Config
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parsing config %s: %w", cfgPath, err)
		}
		if cfg.Address == "" {
			cfg.Address = defaultAddress
		}
		if cfg.Token == "" {
			return nil, fmt.Errorf("config %s has empty token — delete the file and restart to regenerate", cfgPath)
		}
		return &cfg, nil
	}

	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	return create()
}

func create() (*Config, error) {
	dir, err := Dir()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(dir, dirPermissions); err != nil {
		return nil, fmt.Errorf("creating config directory: %w", err)
	}
	if err := os.Chmod(dir, dirPermissions); err != nil {
		return nil, fmt.Errorf("setting config directory permissions: %w", err)
	}

	token, err := auth.Generate()
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Address: defaultAddress,
		Token:   token,
	}

	cfgPath, err := path()
	if err != nil {
		return nil, err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(cfgPath, data, filePermissions); err != nil {
		return nil, fmt.Errorf("writing config: %w", err)
	}

	return cfg, nil
}

func Save(cfg *Config) error {
	cfgPath, err := path()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	return os.WriteFile(cfgPath, data, filePermissions)
}
