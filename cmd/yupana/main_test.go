package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunRejectsUnknownCommands(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "no command", args: []string{"yupana"}},
		{name: "unknown command", args: []string{"yupana", "deploy"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stderr bytes.Buffer
			code := run(context.Background(), tt.args, func(string) string { return "" }, &stderr)
			if code != 2 {
				t.Fatalf("exit code = %d, want 2", code)
			}
			if !strings.Contains(stderr.String(), "usage: yupana <api|worker|bot|migrate|healthcheck>") {
				t.Fatalf("stderr = %q, want usage line", stderr.String())
			}
		})
	}
}

func TestRunReportsInvalidConfig(t *testing.T) {
	var stderr bytes.Buffer
	code := run(context.Background(), []string{"yupana", "api"}, func(string) string { return "" }, &stderr)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "YUPANA_DATABASE_URL is required") {
		t.Fatalf("stderr = %q, want the config error", stderr.String())
	}
}

func TestHealthcheck(t *testing.T) {
	healthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	defer healthy.Close()
	unhealthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer unhealthy.Close()

	tests := []struct {
		name string
		url  string
		want int
	}{
		{name: "200 means healthy", url: healthy.URL, want: 0},
		{name: "503 means unhealthy", url: unhealthy.URL, want: 1},
		{name: "unreachable means unhealthy", url: "http://127.0.0.1:1/healthz", want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stderr bytes.Buffer
			code := run(context.Background(), []string{"yupana", "healthcheck", tt.url}, func(string) string { return "" }, &stderr)
			if code != tt.want {
				t.Fatalf("exit code = %d, want %d (stderr %q)", code, tt.want, stderr.String())
			}
		})
	}
}

func TestMigrateRequiresOwnerURLAndAppPassword(t *testing.T) {
	var stderr bytes.Buffer
	code := run(context.Background(), []string{"yupana", "migrate"}, func(string) string { return "" }, &stderr)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	for _, want := range []string{"YUPANA_MIGRATE_DATABASE_URL is required", "YUPANA_APP_DB_PASSWORD is required"} {
		if !strings.Contains(stderr.String(), want) {
			t.Errorf("stderr = %q, want it to contain %q", stderr.String(), want)
		}
	}
}
