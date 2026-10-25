// Package bot wires the messaging-channel process (Telegram first).
package bot

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/Jaed69/yupana/internal/platform/config"
	"github.com/Jaed69/yupana/internal/platform/httpx"
)

// Run handles chat updates until ctx is cancelled. It serves only /healthz.
// Channel adapters arrive with the ingestion phase.
func Run(ctx context.Context, cfg config.Config, logger *slog.Logger) error {
	if cfg.TelegramToken == "" {
		logger.Warn("YUPANA_TELEGRAM_TOKEN is not set; no channel will be polled")
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	return httpx.Serve(ctx, cfg.HTTPAddr, mux, logger)
}
