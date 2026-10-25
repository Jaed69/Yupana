// Command yupana is the single binary behind every process of the platform.
// The same image runs `yupana api`, `yupana worker` or `yupana bot`.
package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Jaed69/yupana/internal/app/api"
	"github.com/Jaed69/yupana/internal/app/bot"
	"github.com/Jaed69/yupana/internal/app/worker"
	"github.com/Jaed69/yupana/internal/platform/config"
)

const usage = "usage: yupana <api|worker|bot|healthcheck>"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	code := run(ctx, os.Args, os.Getenv, os.Stderr)
	stop()
	os.Exit(code)
}

// run returns the process exit code: 0 ok, 1 runtime failure, 2 bad usage.
func run(ctx context.Context, args []string, getenv func(string) string, stderr io.Writer) int {
	if len(args) < 2 {
		fmt.Fprintln(stderr, usage)
		return 2
	}
	cmd := args[1]
	if cmd == "healthcheck" {
		if len(args) != 3 {
			fmt.Fprintln(stderr, "usage: yupana healthcheck <url>")
			return 2
		}
		return healthcheck(ctx, args[2], stderr)
	}

	var start func(context.Context, config.Config, *slog.Logger) error
	switch cmd {
	case "api":
		start = api.Run
	case "worker":
		start = worker.Run
	case "bot":
		start = bot.Run
	default:
		fmt.Fprintln(stderr, usage)
		return 2
	}

	cfg, err := config.Load(getenv)
	if err != nil {
		fmt.Fprintln(stderr, "invalid configuration:", err)
		return 1
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel})).With("process", cmd)
	logger.Info("starting", "config", cfg.String())
	if err := start(ctx, cfg, logger); err != nil {
		logger.Error("stopped with error", "err", err)
		return 1
	}
	logger.Info("stopped")
	return 0
}

// healthcheck backs the container HEALTHCHECK: the runtime image has no curl.
func healthcheck(ctx context.Context, url string, stderr io.Writer) int {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintln(stderr, "unhealthy:", resp.Status)
		return 1
	}
	return 0
}
