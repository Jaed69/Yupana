// Package worker wires the background job process.
package worker

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/Jaed69/yupana/internal/platform/config"
	"github.com/Jaed69/yupana/internal/platform/httpx"
)

// Run processes background jobs until ctx is cancelled. It serves only /healthz.
func Run(ctx context.Context, cfg config.Config, logger *slog.Logger) error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	return httpx.Serve(ctx, cfg.HTTPAddr, mux, logger)
}
