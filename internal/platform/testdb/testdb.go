// Package testdb gives integration tests a freshly migrated, throwaway database.
package testdb

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/url"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Jaed69/yupana/internal/platform/migrate"
)

// EnvURL names a superuser URL of a disposable server, e.g.
// postgres://postgres:test@localhost:55432/postgres.
const EnvURL = "YUPANA_TEST_DATABASE_URL"

const appPassword = "yupana_app_test"

// migrateLock serializes migrations across test packages: the app role is cluster-wide.
const migrateLock = 727274

// DB holds pools for the owner (bypasses RLS) and the app role (subject to RLS).
type DB struct {
	Owner, App *pgxpool.Pool
	OwnerURL   string
	AppURL     string
}

// New creates and migrates a uniquely named database, dropped at cleanup.
// It is skipped with -short and fails when EnvURL is unset.
func New(t *testing.T) DB {
	t.Helper()
	if testing.Short() {
		t.Skip("integration test: needs Postgres")
	}
	admin := os.Getenv(EnvURL)
	if admin == "" {
		t.Fatalf("%s is not set; start Postgres (see README) or run with -short", EnvURL)
	}
	ctx := context.Background()
	adminConn, err := pgx.Connect(ctx, admin)
	if err != nil {
		t.Fatalf("connect admin: %v", err)
	}
	t.Cleanup(func() { adminConn.Close(context.Background()) })

	name := "yupana_t_" + randomHex(t)
	if _, err := adminConn.Exec(ctx, "CREATE DATABASE "+name); err != nil {
		t.Fatalf("create database: %v", err)
	}
	t.Cleanup(func() {
		if _, err := adminConn.Exec(context.Background(), "DROP DATABASE "+name+" WITH (FORCE)"); err != nil {
			t.Errorf("drop database %s: %v", name, err)
		}
	})

	ownerURL := withDatabase(t, admin, name)
	if _, err := adminConn.Exec(ctx, "SELECT pg_advisory_lock($1)", migrateLock); err != nil {
		t.Fatal(err)
	}
	err = migrate.Up(ctx, ownerURL, appPassword, slog.New(slog.DiscardHandler))
	if _, unlockErr := adminConn.Exec(ctx, "SELECT pg_advisory_unlock($1)", migrateLock); unlockErr != nil {
		t.Fatal(unlockErr)
	}
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}

	appURL := withUser(t, ownerURL, migrate.AppRole, appPassword)
	return DB{Owner: pool(t, ownerURL), App: pool(t, appURL), OwnerURL: ownerURL, AppURL: appURL}
}

func pool(t *testing.T, u string) *pgxpool.Pool {
	t.Helper()
	p, err := pgxpool.New(context.Background(), u)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(p.Close)
	return p
}

func withDatabase(t *testing.T, raw, name string) string {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	u.Path = "/" + name
	return u.String()
}

func withUser(t *testing.T, raw, user, password string) string {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	u.User = url.UserPassword(user, password)
	return u.String()
}

func randomHex(t *testing.T) string {
	t.Helper()
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(b)
}
