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

// Load reads configuration in this order:
//  1. ONE_AUTH + ONE_XMLRPC environment variables (backwards compatible)
//  2. ~/.one9s/config file (recommended)
func Load() (Config, error) {
	cfg := Config{}

	// --- 1. Environment variables ---
	if ep := os.Getenv("ONE_XMLRPC"); ep != "" {
		cfg.Endpoint = ep
	}
	if auth := os.Getenv("ONE_AUTH"); auth != "" {
		u, p, err := parseAuthString(auth)
		if err != nil {
			return Config{}, fmt.Errorf("ONE_AUTH: %w", err)
		}
		cfg.User = u
		cfg.Password = p
	}

	// --- 2. ~/.one9s/config ---
	if cfg.User == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			configPath := filepath.Join(home, ".one9s", "config")
			if data, readErr := os.ReadFile(configPath); readErr == nil {
				cfg = parseConfigFile(string(data), cfg)
			}
		}
	}

	// --- Defaults ---
	if cfg.Endpoint == "" {
		cfg.Endpoint = "http://localhost:2633/RPC2"
	}

	// --- Validation ---
	if cfg.User == "" || cfg.Password == "" {
		home, _ := os.UserHomeDir()
		configPath := filepath.Join(home, ".one9s", "config")
		return Config{}, fmt.Errorf(
			"no OpenNebula credentials found\n\n"+
				"Create ~/.one9s/config:\n"+
				"  mkdir -p ~/.one9s\n"+
				"  echo 'ONE_XMLRPC=http://opennebula:2633/RPC2' > %s\n"+
				"  echo 'ONE_AUTH=oneadmin:password' >> %s\n\n"+
				"Or set environment variables:\n"+
				"  export ONE_AUTH=\"oneadmin:password\"\n"+
				"  export ONE_XMLRPC=\"http://opennebula:2633/RPC2\"",
			configPath, configPath,
		)
	}

	return cfg, nil
}

// parseConfigFile parses KEY=VALUE lines from the config file.
func parseConfigFile(data string, cfg Config) Config {
	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		switch key {
		case "ONE_XMLRPC":
			if cfg.Endpoint == "" {
				cfg.Endpoint = value
			}
		case "ONE_AUTH":
			if cfg.User == "" {
				u, p, err := parseAuthString(value)
				if err == nil {
					cfg.User = u
					cfg.Password = p
				}
			}
		}
	}
	return cfg
}

// parseAuthString parses "user:password" format.
func parseAuthString(auth string) (user, pass string, err error) {
	if idx := strings.IndexByte(auth, ':'); idx >= 0 {
		return auth[:idx], auth[idx+1:], nil
	}
	return "", "", fmt.Errorf("expected user:password format")
}
