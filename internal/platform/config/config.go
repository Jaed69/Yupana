// Package config loads process configuration from environment variables.
package config

import (
	"errors"
	"fmt"
	"log/slog"
)

// Config is shared by every entrypoint (api, worker, bot, migrate).
type Config struct {
	// Env is "dev" or "prod". Dev-only conveniences are gated on it.
	Env string
	// HTTPAddr is where the process serves HTTP (api routes, or /healthz only).
	HTTPAddr string
	// DatabaseURL connects as the non-owner application role (subject to RLS).
	DatabaseURL string
	// MigrateDatabaseURL connects as the schema owner; only `migrate` uses it.
	MigrateDatabaseURL string
	LogLevel           slog.Level
	TelegramToken      string
}

// Load reads the configuration through getenv (os.Getenv in production).
func Load(getenv func(string) string) (Config, error) {
	c := Config{
		Env:                withDefault(getenv("YUPANA_ENV"), "prod"),
		HTTPAddr:           withDefault(getenv("YUPANA_HTTP_ADDR"), ":8080"),
		DatabaseURL:        getenv("YUPANA_DATABASE_URL"),
		MigrateDatabaseURL: getenv("YUPANA_MIGRATE_DATABASE_URL"),
		TelegramToken:      getenv("YUPANA_TELEGRAM_TOKEN"),
	}
	var errs []error
	if c.Env != "dev" && c.Env != "prod" {
		errs = append(errs, fmt.Errorf(`YUPANA_ENV must be "dev" or "prod", got %q`, c.Env))
	}
	if c.DatabaseURL == "" {
		errs = append(errs, errors.New("YUPANA_DATABASE_URL is required"))
	}
	if err := c.LogLevel.UnmarshalText([]byte(withDefault(getenv("YUPANA_LOG_LEVEL"), "info"))); err != nil {
		errs = append(errs, fmt.Errorf("YUPANA_LOG_LEVEL: %w", err))
	}
	return c, errors.Join(errs...)
}

// String keeps credentials out of logs.
func (c Config) String() string {
	return fmt.Sprintf("Config{Env:%s HTTPAddr:%s LogLevel:%s DatabaseURL:%s MigrateDatabaseURL:%s TelegramToken:%s}",
		c.Env, c.HTTPAddr, c.LogLevel, redact(c.DatabaseURL), redact(c.MigrateDatabaseURL), redact(c.TelegramToken))
}

func redact(s string) string {
	if s == "" {
		return "unset"
	}
	return "set"
}

func withDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
