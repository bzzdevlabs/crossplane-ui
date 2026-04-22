// Package config loads the auth service configuration from the environment.
package config

import (
	"errors"
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

	// Namespace is the namespace in which the auth service runs. Secrets and
	// Dex storage CRs read/written by the controller live here. Defaults to
	// the content of /var/run/secrets/kubernetes.io/serviceaccount/namespace
	// or the POD_NAMESPACE environment variable.
	Namespace string

	// BootstrapAdminUsername is the username of the administrator created on
	// first startup when no User CR exists yet.
	BootstrapAdminUsername string
	// BootstrapAdminPasswordSecret is the name of the Kubernetes Secret that
	// holds the bootstrap administrator password. The Secret must live in the
	// same namespace as the auth service and expose a "password" key.
	BootstrapAdminPasswordSecret string

	// OIDC client configuration — used at startup to (re)materialise the
	// OAuth2Client Dex custom resource the gateway authenticates against.
	OIDCClientID           string
	OIDCClientName         string
	OIDCClientSecretSecret string
	OIDCClientSecretKey    string
	OIDCRedirectURIs       []string

	// LeaderElection toggles controller-runtime leader election. Defaults to
	// true; disable only when running a single-replica controller during dev.
	LeaderElection bool
	// MetricsAddr is where the controller-runtime metrics server binds. An
	// empty value disables the metrics server.
	MetricsAddr string
	// HealthProbeAddr is where the controller-runtime health endpoints bind.
	// Leave empty to rely solely on our own /healthz and /readyz.
	HealthProbeAddr string
}

// Load reads the configuration from the process environment.
func Load() (*Config, error) {
	cfg := &Config{
		HTTPAddr:                     getString("HTTP_ADDR", ":8081"),
		ReadHeaderTimeout:            getDuration("HTTP_READ_HEADER_TIMEOUT", 10*time.Second),
		LogLevel:                     strings.ToLower(getString("LOG_LEVEL", "info")),
		LogFormat:                    strings.ToLower(getString("LOG_FORMAT", "json")),
		KubeconfigPath:               os.Getenv("KUBECONFIG"),
		Namespace:                    resolveNamespace(),
		BootstrapAdminUsername:       getString("BOOTSTRAP_ADMIN_USERNAME", "admin"),
		BootstrapAdminPasswordSecret: getString("BOOTSTRAP_ADMIN_PASSWORD_SECRET", "crossplane-ui-bootstrap-admin"),
		OIDCClientID:                 getString("OIDC_CLIENT_ID", ""),
		OIDCClientName:               getString("OIDC_CLIENT_NAME", "crossplane-ui"),
		OIDCClientSecretSecret:       getString("OIDC_CLIENT_SECRET_NAME", ""),
		OIDCClientSecretKey:          getString("OIDC_CLIENT_SECRET_KEY", "clientSecret"),
		OIDCRedirectURIs:             splitList(getString("OIDC_REDIRECT_URIS", "")),
		LeaderElection:               getBool("LEADER_ELECTION", true),
		MetricsAddr:                  getString("METRICS_ADDR", ":8082"),
		HealthProbeAddr:              getString("HEALTH_PROBE_ADDR", ""),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// resolveNamespace reads the pod namespace. The service account token file is
// the authoritative source in-cluster; POD_NAMESPACE is the escape hatch for
// local runs.
func resolveNamespace() string {
	if v := os.Getenv("POD_NAMESPACE"); v != "" {
		return v
	}
	if b, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		ns := strings.TrimSpace(string(b))
		if ns != "" {
			return ns
		}
	}
	return getString("NAMESPACE", "crossplane-ui")
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
		return errors.New("HTTP_ADDR must not be empty")
	}
	return nil
}

func getBool(key string, def bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return def
	}
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

func splitList(s string) []string {
	fields := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n'
	})
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if f = strings.TrimSpace(f); f != "" {
			out = append(out, f)
		}
	}
	return out
}
