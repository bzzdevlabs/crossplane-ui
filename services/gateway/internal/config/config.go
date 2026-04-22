// Package config loads the gateway configuration from environment variables.
//
// No third-party dependency is used on purpose: the scaffold is kept
// zero-dependency so that it builds in air-gapped environments without
// GOPROXY access. A richer parser (e.g. caarlos0/env) will replace this if
// the surface grows significantly.
package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Config holds the runtime configuration of the gateway.
type Config struct {
	// HTTPAddr is the listening address of the public HTTP server.
	HTTPAddr string
	// ReadHeaderTimeout is the maximum duration for reading request headers.
	ReadHeaderTimeout time.Duration
	// LogLevel is one of debug, info, warn, error.
	LogLevel string
	// LogFormat is one of json, text.
	LogFormat string
	// KubeconfigPath is the optional path to a kubeconfig file. When empty,
	// in-cluster configuration is used.
	KubeconfigPath string

	// OIDCIssuerURL is the expected `iss` claim in tokens. When empty the
	// gateway runs in dev-passthrough mode (no auth), which is fine for
	// local development only.
	OIDCIssuerURL string
	// OIDCDiscoveryURL is where the gateway fetches the OIDC discovery
	// document. Defaults to OIDCIssuerURL when empty. Splitting the two
	// matters in dev setups where Dex advertises one hostname to the
	// browser and another to the gateway.
	OIDCDiscoveryURL string
	// OIDCClientID is the client identifier registered in the IdP.
	OIDCClientID string
	// OIDCSkipIssuerCheck disables the iss check; dev-only escape hatch.
	OIDCSkipIssuerCheck bool

	// AuthServiceURL is the base URL of the companion auth service.
	AuthServiceURL string

	// CORSAllowedOrigins is the list of Origin values accepted by the
	// gateway; use a single "*" to allow any origin (dev only).
	CORSAllowedOrigins []string
}

// Load reads the configuration from the process environment and returns a
// validated Config. Unknown keys are ignored.
func Load() (*Config, error) {
	cfg := &Config{
		HTTPAddr:            getString("HTTP_ADDR", ":8080"),
		ReadHeaderTimeout:   getDuration("HTTP_READ_HEADER_TIMEOUT", 10*time.Second),
		LogLevel:            strings.ToLower(getString("LOG_LEVEL", "info")),
		LogFormat:           strings.ToLower(getString("LOG_FORMAT", "json")),
		KubeconfigPath:      os.Getenv("KUBECONFIG"),
		OIDCIssuerURL:       os.Getenv("OIDC_ISSUER_URL"),
		OIDCDiscoveryURL:    os.Getenv("OIDC_DISCOVERY_URL"),
		OIDCClientID:        os.Getenv("OIDC_CLIENT_ID"),
		OIDCSkipIssuerCheck: getBool("OIDC_SKIP_ISSUER_CHECK", false),
		AuthServiceURL:      getString("AUTH_SERVICE_URL", "http://auth.crossplane-ui.svc.cluster.local:8081"),
		CORSAllowedOrigins:  splitCSV(getString("CORS_ALLOWED_ORIGINS", "http://localhost:5173")),
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
	if c.OIDCIssuerURL != "" && c.OIDCClientID == "" {
		return fmt.Errorf("OIDC_CLIENT_ID must be set when OIDC_ISSUER_URL is configured")
	}
	return nil
}

// AuthEnabled reports whether OIDC middleware should be mounted.
func (c *Config) AuthEnabled() bool { return c.OIDCIssuerURL != "" }

// EffectiveDiscoveryURL returns the URL to fetch OIDC discovery from.
func (c *Config) EffectiveDiscoveryURL() string {
	if c.OIDCDiscoveryURL != "" {
		return c.OIDCDiscoveryURL
	}
	return c.OIDCIssuerURL
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

func getBool(key string, def bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return def
	}
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
