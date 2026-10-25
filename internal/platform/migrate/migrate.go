// Package migrate applies the embedded SQL migrations as the schema owner.
package migrate

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" database/sql driver
	"github.com/pressly/goose/v3"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"

	"github.com/Jaed69/yupana/internal/platform/queue"
)

//go:embed sql/*.sql
var files embed.FS

// AppRole is the login role the processes use; it is subject to RLS.
const AppRole = "yupana_app"

// Up applies pending migrations through ownerURL and sets the password of
// AppRole, so the secret lives in the environment, never in a migration file.
func Up(ctx context.Context, ownerURL, appPassword string, logger *slog.Logger) error {
	if appPassword == "" {
		return errors.New("app role password is required")
	}
	sqlFS, err := fs.Sub(files, "sql")
	if err != nil {
		return err
	}
	conn, err := sql.Open("pgx", ownerURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	provider, err := goose.NewProvider(goose.DialectPostgres, conn, sqlFS)
	if err != nil {
		return err
	}
	results, err := provider.Up(ctx)
	if err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	for _, r := range results {
		logger.Info("migration applied", "version", r.Source.Version, "file", r.Source.Path, "duration", r.Duration)
	}

	// DDL takes no bind parameters: let Postgres quote the literal.
	var stmt string
	err = conn.QueryRowContext(ctx, `SELECT format('ALTER ROLE %I WITH LOGIN PASSWORD %L', $1::text, $2::text)`,
		AppRole, appPassword).Scan(&stmt)
	if err != nil {
		return err
	}
	if _, err := conn.ExecContext(ctx, stmt); err != nil {
		return fmt.Errorf("set %s password: %w", pgx.Identifier{AppRole}.Sanitize(), err)
	}
	return upQueue(ctx, ownerURL, logger)
}

// upQueue applies River's own migrations in queue.Schema and grants the app
// role what a River client needs: job DML, but read-only migration history.
func upQueue(ctx context.Context, ownerURL string, logger *slog.Logger) error {
	pool, err := pgxpool.New(ctx, ownerURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if _, err := pool.Exec(ctx, `CREATE SCHEMA IF NOT EXISTS `+queue.Schema); err != nil {
		return err
	}
	migrator, err := rivermigrate.New(riverpgxv5.New(pool), &rivermigrate.Config{Schema: queue.Schema, Logger: logger})
	if err != nil {
		return err
	}
	res, err := migrator.Migrate(ctx, rivermigrate.DirectionUp, nil)
	if err != nil {
		return fmt.Errorf("apply queue migrations: %w", err)
	}
	for _, v := range res.Versions {
		logger.Info("queue migration applied", "version", v.Version, "duration", v.Duration)
	}

	role := pgx.Identifier{AppRole}.Sanitize()
	_, err = pool.Exec(ctx, fmt.Sprintf(`
		GRANT USAGE ON SCHEMA %[1]s TO %[2]s;
		GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA %[1]s TO %[2]s;
		REVOKE INSERT, UPDATE, DELETE ON %[1]s.river_migration FROM %[2]s;
		GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA %[1]s TO %[2]s;`, queue.Schema, role))
	if err != nil {
		return fmt.Errorf("grant queue access: %w", err)
	}
	return nil
}
