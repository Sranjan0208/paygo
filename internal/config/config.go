package config

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Config holds all runtime configuration, read once at startup.
type Config struct {
	Env             string        // dev | staging | prod
	HTTPAddr        string        // default ":8080"
	DatabaseURL     string        // required
	ShutdownTimeout time.Duration // default 15s
}

// Load builds a Config using getenv (os.Getenv in production, a fake in tests).
// It reports every problem at once and returns a zero Config if there are any.
func Load(getenv func(string) string) (Config, error) {
	get := func(key string) string {
		return strings.TrimSpace(getenv(key))
	}

	var cfg Config
	var errs []error

	// APP_ENV
	cfg.Env = get("APP_ENV")
	if cfg.Env == "" {
		cfg.Env = "dev"
	}
	switch cfg.Env {
	case "dev", "staging", "prod":
		// valid
	default:
		errs = append(errs, fmt.Errorf("invalid APP_ENV %q: must be dev, staging or prod", cfg.Env))
	}

	// HTTP_ADDR
	cfg.HTTPAddr = get("HTTP_ADDR")
	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = ":8080"
	}

	// DATABASE_URL
	cfg.DatabaseURL = get("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		errs = append(errs, errors.New("DATABASE_URL is required"))
	}

	// SHUTDOWN_TIMEOUT
	raw := get("SHUTDOWN_TIMEOUT")
	if raw == "" {
		raw = "15s"
	}
	duration, err := time.ParseDuration(raw)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid SHUTDOWN_TIMEOUT %q: %w", raw, err))
	} else if duration <= 0 {
		errs = append(errs, fmt.Errorf("SHUTDOWN_TIMEOUT must be positive, got %s", duration))
	} else {
		cfg.ShutdownTimeout = duration
	}

	if len(errs) > 0 {
		return Config{}, errors.Join(errs...)
	}
	return cfg, nil
}
