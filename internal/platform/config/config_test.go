package config_test

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/Jaed69/yupana/internal/platform/config"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		want    config.Config
		wantErr string
	}{
		{
			name: "defaults when only the database URL is set",
			env:  map[string]string{"YUPANA_DATABASE_URL": "postgres://app@db/yupana"},
			want: config.Config{
				Env:         "prod",
				HTTPAddr:    ":8080",
				DatabaseURL: "postgres://app@db/yupana",
				LogLevel:    slog.LevelInfo,
			},
		},
		{
			name: "explicit values override defaults",
			env: map[string]string{
				"YUPANA_ENV":                  "dev",
				"YUPANA_HTTP_ADDR":            ":9000",
				"YUPANA_DATABASE_URL":         "postgres://app@db/yupana",
				"YUPANA_MIGRATE_DATABASE_URL": "postgres://owner@db/yupana",
				"YUPANA_LOG_LEVEL":            "debug",
				"YUPANA_TELEGRAM_TOKEN":       "123:abc",
			},
			want: config.Config{
				Env:                "dev",
				HTTPAddr:           ":9000",
				DatabaseURL:        "postgres://app@db/yupana",
				MigrateDatabaseURL: "postgres://owner@db/yupana",
				LogLevel:           slog.LevelDebug,
				TelegramToken:      "123:abc",
			},
		},
		{
			name:    "database URL is required",
			env:     map[string]string{},
			wantErr: "YUPANA_DATABASE_URL is required",
		},
		{
			name:    "unknown environment is rejected",
			env:     map[string]string{"YUPANA_DATABASE_URL": "x", "YUPANA_ENV": "staging"},
			wantErr: `YUPANA_ENV must be "dev" or "prod"`,
		},
		{
			name:    "invalid log level is rejected",
			env:     map[string]string{"YUPANA_DATABASE_URL": "x", "YUPANA_LOG_LEVEL": "loud"},
			wantErr: "YUPANA_LOG_LEVEL",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := config.Load(func(k string) string { return tt.env[k] })
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("Load() error = %v, want it to contain %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Load() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("Load() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestConfigStringRedactsSecrets(t *testing.T) {
	c := config.Config{
		DatabaseURL:        "postgres://app:s3cret@db/yupana",
		MigrateDatabaseURL: "postgres://owner:0wner@db/yupana",
		TelegramToken:      "123:tok3n",
	}
	s := c.String()
	for _, secret := range []string{"s3cret", "0wner", "tok3n"} {
		if strings.Contains(s, secret) {
			t.Fatalf("String() leaks %q: %s", secret, s)
		}
	}
}
