// Package api wires the HTTP API process.
package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Jaed69/yupana/internal/modules/ingest"
	"github.com/Jaed69/yupana/internal/platform/config"
	"github.com/Jaed69/yupana/internal/platform/db"
	"github.com/Jaed69/yupana/internal/platform/httpx"
	"github.com/Jaed69/yupana/internal/platform/modules"
	"github.com/Jaed69/yupana/internal/platform/tenant"
)

// Run serves the API until ctx is cancelled.
func Run(ctx context.Context, cfg config.Config, logger *slog.Logger) error {
	pool, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	return httpx.Serve(ctx, cfg.HTTPAddr, newHandler(cfg, pool), logger)
}

func newHandler(cfg config.Config, pool *pgxpool.Pool) http.Handler {
	// Until authentication lands, only dev may pick a tenant (through a header).
	var res tenant.Resolver = tenant.DenyResolver{}
	if cfg.Env == "dev" {
		res = tenant.HeaderResolver{}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	modules.Mount(mux, pool, res, ingest.Module{})
	return mux
}
