package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds OpenNebula connection configuration.
type Config struct {
	Endpoint string
	User     string
	Password string
}

// Load reads ONE_AUTH / ONE_XMLRPC env vars or falls back to ~/.one/one_auth.
func Load() (Config, error) {
	cfg := Config{
		Endpoint: os.Getenv("ONE_XMLRPC"),
	}

	// Try ONE_AUTH env var first (format: "user:password" or path to file).
	if auth := os.Getenv("ONE_AUTH"); auth != "" {
		if strings.Contains(auth, ":") && !strings.Contains(auth, string(filepath.Separator)) {
			parts := strings.SplitN(auth, ":", 2)
			cfg.User = parts[0]
			cfg.Password = parts[1]
		} else {
			data, err := os.ReadFile(auth)
			if err != nil {
				return Config{}, fmt.Errorf("read ONE_AUTH file %s: %w", auth, err)
			}
			cfg.User, cfg.Password, err = parseAuthFile(data)
			if err != nil {
				return Config{}, err
			}
		}
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return Config{}, fmt.Errorf("home dir: %w", err)
		}
		authPath := filepath.Join(home, ".one", "one_auth")
		data, err := os.ReadFile(authPath)
		if err != nil {
			return Config{}, fmt.Errorf("read %s: %w", authPath, err)
		}
		cfg.User, cfg.Password, err = parseAuthFile(data)
		if err != nil {
			return Config{}, err
		}
	}

	if cfg.Endpoint == "" {
		cfg.Endpoint = "http://localhost:2633/RPC2"
	}

	if cfg.User == "" || cfg.Password == "" {
		return Config{}, fmt.Errorf("missing OpenNebula credentials (set ONE_AUTH or ~/.one/one_auth)")
	}

	return cfg, nil
}

func parseAuthFile(data []byte) (user, pass string, err error) {
	raw := strings.TrimSpace(string(data))
	if idx := strings.IndexByte(raw, ':'); idx >= 0 {
		return raw[:idx], raw[idx+1:], nil
	}
	return "", "", fmt.Errorf("invalid auth file format (expected user:password)")
}

// Session returns the XML-RPC session string "user:password".
func (c Config) Session() string {
	return c.User + ":" + c.Password
}
