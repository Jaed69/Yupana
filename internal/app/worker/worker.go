// Package worker wires the background job process.
package worker

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/riverqueue/river"

	"github.com/Jaed69/yupana/internal/platform/config"
	"github.com/Jaed69/yupana/internal/platform/db"
	"github.com/Jaed69/yupana/internal/platform/httpx"
	"github.com/Jaed69/yupana/internal/platform/queue"
)

// Run works background jobs until ctx is cancelled. It serves only /healthz.
func Run(ctx context.Context, cfg config.Config, logger *slog.Logger) error {
	pool, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	workers := river.NewWorkers()
	river.AddWorker(workers, &queue.PingWorker{})
	client, err := queue.NewClient(pool, workers, logger)
	if err != nil {
		return err
	}
	if err := client.Start(ctx); err != nil {
		return err
	}
	defer func() {
		// Let running jobs finish; River returns them to the queue if they don't.
		stopCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := client.Stop(stopCtx); err != nil {
			logger.Error("stop job client", "err", err)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	return httpx.Serve(ctx, cfg.HTTPAddr, mux, logger)
}
