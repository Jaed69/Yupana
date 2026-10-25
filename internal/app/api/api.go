// Package api wires the HTTP API process.
package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/Jaed69/yupana/internal/platform/config"
	"github.com/Jaed69/yupana/internal/platform/httpx"
)

// Run serves the API until ctx is cancelled.
func Run(ctx context.Context, cfg config.Config, logger *slog.Logger) error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	return httpx.Serve(ctx, cfg.HTTPAddr, mux, logger)
}
