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
	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" database/sql driver
	"github.com/pressly/goose/v3"
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
	return nil
}
