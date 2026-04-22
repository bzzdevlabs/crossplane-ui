// Package config loads the auth service configuration from the environment.
package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Config holds the runtime configuration of the auth service.
type Config struct {
	HTTPAddr          string
	ReadHeaderTimeout time.Duration
	LogLevel          string
	LogFormat         string
	KubeconfigPath    string

	// BootstrapAdminUsername is the username of the administrator created on
	// first startup when no User CR exists yet.
	BootstrapAdminUsername string
	// BootstrapAdminPasswordSecret is the name of the Kubernetes Secret that
	// holds the bootstrap administrator password. The Secret must live in the
	// same namespace as the auth service and expose a "password" key.
	BootstrapAdminPasswordSecret string

	// DexConfigMapName is the ConfigMap that holds the Dex configuration. The
	// auth service rewrites the staticPasswords section whenever User CRs
	// change.
	DexConfigMapName string
}

// Load reads the configuration from the process environment.
func Load() (*Config, error) {
	cfg := &Config{
		HTTPAddr:                     getString("HTTP_ADDR", ":8081"),
		ReadHeaderTimeout:            getDuration("HTTP_READ_HEADER_TIMEOUT", 10*time.Second),
		LogLevel:                     strings.ToLower(getString("LOG_LEVEL", "info")),
		LogFormat:                    strings.ToLower(getString("LOG_FORMAT", "json")),
		KubeconfigPath:               os.Getenv("KUBECONFIG"),
		BootstrapAdminUsername:       getString("BOOTSTRAP_ADMIN_USERNAME", "admin"),
		BootstrapAdminPasswordSecret: getString("BOOTSTRAP_ADMIN_PASSWORD_SECRET", "crossplane-ui-bootstrap-admin"),
		DexConfigMapName:             getString("DEX_CONFIGMAP_NAME", "crossplane-ui-dex-config"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	switch c.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("invalid LOG_LEVEL %q (allowed: debug, info, warn, error)", c.LogLevel)
	}
	switch c.LogFormat {
	case "json", "text":
	default:
		return fmt.Errorf("invalid LOG_FORMAT %q (allowed: json, text)", c.LogFormat)
	}
	if c.HTTPAddr == "" {
		return fmt.Errorf("HTTP_ADDR must not be empty")
	}
	return nil
}

func getString(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func getDuration(key string, def time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
