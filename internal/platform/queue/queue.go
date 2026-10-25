// Package queue builds River clients over the app pool.
//
// Jobs live in Postgres (schema river), so a job inserted with InsertTx
// inside a business transaction exists only if that transaction commits.
// Job args must carry the tenant ID; workers scope their reads and writes
// with db.WithTenant, never with the pool directly.
package queue

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

// Schema holds River's tables, apart from business data.
const Schema = "river"

// NewClient returns a River client. With nil workers it only inserts (the API
// process); with workers it also works the default queue (the worker process).
func NewClient(pool *pgxpool.Pool, workers *river.Workers, logger *slog.Logger) (*river.Client[pgx.Tx], error) {
	cfg := &river.Config{
		Schema: Schema,
		Logger: logger,
		// REINDEX requires owning the tables, which the app role never does.
		ReindexerSchedule: river.NeverSchedule(),
	}
	if workers != nil {
		cfg.Workers = workers
		cfg.Queues = map[string]river.QueueConfig{river.QueueDefault: {MaxWorkers: 10}}
	}
	return river.NewClient(riverpgxv5.New(pool), cfg)
}

// PingArgs is a platform job that does no business work. It proves the
// insert-commit-work path end to end until modules register real jobs.
type PingArgs struct {
	TenantID uuid.UUID `json:"tenant_id"`
}

func (PingArgs) Kind() string { return "platform.ping" }

// PingWorker logs each ping it works.
type PingWorker struct {
	river.WorkerDefaults[PingArgs]
}

func (PingWorker) Work(ctx context.Context, job *river.Job[PingArgs]) error {
	slog.InfoContext(ctx, "ping worked", "job_id", job.ID, "tenant_id", job.Args.TenantID)
	return nil
}
