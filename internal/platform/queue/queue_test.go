package queue_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"

	"github.com/Jaed69/yupana/internal/platform/db"
	"github.com/Jaed69/yupana/internal/platform/queue"
	"github.com/Jaed69/yupana/internal/platform/testdb"
)

type probeArgs struct {
	TenantID uuid.UUID `json:"tenant_id"`
}

func (probeArgs) Kind() string { return "test.probe" }

// probeWorker reads tenant data inside the job's tenant scope and reports what it saw.
type probeWorker struct {
	river.WorkerDefaults[probeArgs]
	app  *testdb.DB
	seen chan int
}

func (w *probeWorker) Work(ctx context.Context, job *river.Job[probeArgs]) error {
	return db.WithTenant(ctx, w.app.App, job.Args.TenantID, func(tx pgx.Tx) error {
		var n int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM platform.tenant_modules`).Scan(&n); err != nil {
			return err
		}
		w.seen <- n
		return nil
	})
}

func jobCount(t *testing.T, d testdb.DB) int {
	t.Helper()
	var n int
	if err := d.Owner.QueryRow(context.Background(), `SELECT count(*) FROM river.river_job`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestJobsAreEnqueuedOnlyWhenTheBusinessTransactionCommits(t *testing.T) {
	d := testdb.New(t)
	ctx := context.Background()
	a := uuid.New()
	if _, err := d.Owner.Exec(ctx, `INSERT INTO platform.tenants (id, name) VALUES ($1, 'A')`, a); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Owner.Exec(ctx, `INSERT INTO platform.tenant_modules (tenant_id, module) VALUES ($1, 'ingest'), ($1, 'reports')`, a); err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.DiscardHandler)

	// The API process inserts without working jobs.
	inserter, err := queue.NewClient(d.App, nil, logger)
	if err != nil {
		t.Fatal(err)
	}

	errBusiness := errors.New("business rule failed")
	err = db.WithTenant(ctx, d.App, a, func(tx pgx.Tx) error {
		if _, err := inserter.InsertTx(ctx, tx, probeArgs{TenantID: a}, nil); err != nil {
			return err
		}
		return errBusiness
	})
	if !errors.Is(err, errBusiness) {
		t.Fatalf("err = %v, want the business error", err)
	}
	if n := jobCount(t, d); n != 0 {
		t.Fatalf("after rollback there are %d jobs, want 0", n)
	}

	err = db.WithTenant(ctx, d.App, a, func(tx pgx.Tx) error {
		_, err := inserter.InsertTx(ctx, tx, probeArgs{TenantID: a}, nil)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if n := jobCount(t, d); n != 1 {
		t.Fatalf("after commit there are %d jobs, want 1", n)
	}

	// The worker process picks it up, connected as the RLS-bound app role.
	w := &probeWorker{app: &d, seen: make(chan int, 1)}
	workers := river.NewWorkers()
	river.AddWorker(workers, w)
	worker, err := queue.NewClient(d.App, workers, logger)
	if err != nil {
		t.Fatal(err)
	}
	if err := worker.Start(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		worker.Stop(stopCtx)
	})

	select {
	case n := <-w.seen:
		if n != 2 {
			t.Fatalf("worker saw %d tenant_modules rows in the job's tenant scope, want 2", n)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("committed job was not worked")
	}
}

func TestAppRoleCannotRewriteQueueMigrations(t *testing.T) {
	d := testdb.New(t)
	_, err := d.App.Exec(context.Background(), `DELETE FROM river.river_migration`)
	if err == nil {
		t.Fatal("app role deleted river migrations, want permission denied")
	}
}
