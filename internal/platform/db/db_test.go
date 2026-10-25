package db_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Jaed69/yupana/internal/platform/db"
	"github.com/Jaed69/yupana/internal/platform/testdb"
)

// fixture seeds two tenants as the owner and returns pools for both roles.
type fixture struct {
	owner, app *pgxpool.Pool
	a, b       uuid.UUID
}

func setup(t *testing.T) fixture {
	t.Helper()
	d := testdb.New(t)
	f := fixture{owner: d.Owner, app: d.App, a: uuid.New(), b: uuid.New()}
	mustExec(t, f.owner, `INSERT INTO platform.tenants (id, name) VALUES ($1, 'A'), ($2, 'B')`, f.a, f.b)
	mustExec(t, f.owner, `INSERT INTO platform.tenant_modules (tenant_id, module) VALUES ($1, 'ingest'), ($1, 'reports'), ($2, 'ingest')`, f.a, f.b)
	// A module-owned table protected through the shared helper, like every module migration does.
	mustExec(t, f.owner, `CREATE TABLE platform.rls_probe (id serial PRIMARY KEY, tenant_id uuid NOT NULL, note text)`)
	mustExec(t, f.owner, `SELECT platform.enable_tenant_rls('platform.rls_probe')`)
	mustExec(t, f.owner, `GRANT SELECT, INSERT, UPDATE, DELETE ON platform.rls_probe TO yupana_app`)
	mustExec(t, f.owner, `GRANT USAGE ON SEQUENCE platform.rls_probe_id_seq TO yupana_app`)
	mustExec(t, f.owner, `INSERT INTO platform.rls_probe (tenant_id, note) VALUES ($1, 'a'), ($2, 'b')`, f.a, f.b)
	return f
}

func mustExec(t *testing.T, p *pgxpool.Pool, sql string, args ...any) {
	t.Helper()
	if _, err := p.Exec(context.Background(), sql, args...); err != nil {
		t.Fatalf("exec %q: %v", sql, err)
	}
}

func count(t *testing.T, q interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, table string) int {
	t.Helper()
	var n int
	if err := q.QueryRow(context.Background(), "SELECT count(*) FROM "+table).Scan(&n); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return n
}

func TestWithoutTenantTheAppSeesZeroRows(t *testing.T) {
	f := setup(t)
	for _, table := range []string{"platform.tenants", "platform.tenant_modules", "platform.rls_probe"} {
		if n := count(t, f.app, table); n != 0 {
			t.Errorf("%s: app role without tenant sees %d rows, want 0", table, n)
		}
	}
	if n := count(t, f.owner, "platform.tenant_modules"); n != 3 {
		t.Fatalf("owner sees %d tenant_modules rows, want 3 (fixture broken)", n)
	}
}

func TestTenantSeesOnlyItsOwnRows(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	tests := []struct {
		name    string
		tenant  uuid.UUID
		modules int
	}{
		{name: "tenant A", tenant: f.a, modules: 2},
		{name: "tenant B", tenant: f.b, modules: 1},
		{name: "unknown tenant", tenant: uuid.New(), modules: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.WithTenant(ctx, f.app, tt.tenant, func(tx pgx.Tx) error {
				if n := count(t, tx, "platform.tenant_modules"); n != tt.modules {
					t.Errorf("tenant_modules = %d rows, want %d", n, tt.modules)
				}
				var foreign int
				if err := tx.QueryRow(ctx, `SELECT count(*) FROM platform.tenant_modules WHERE tenant_id <> $1`, tt.tenant).Scan(&foreign); err != nil {
					return err
				}
				if foreign != 0 {
					t.Errorf("sees %d rows of other tenants, want 0", foreign)
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestWritesForAnotherTenantAreRejected(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	tests := []struct {
		name string
		sql  string
	}{
		{name: "insert", sql: `INSERT INTO platform.rls_probe (tenant_id, note) VALUES ($1, 'smuggled')`},
		{name: "update moving an own row", sql: `UPDATE platform.rls_probe SET tenant_id = $1`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.WithTenant(ctx, f.app, f.a, func(tx pgx.Tx) error {
				_, err := tx.Exec(ctx, tt.sql, f.b)
				return err
			})
			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) || pgErr.Code != "42501" {
				t.Fatalf("err = %v, want SQLSTATE 42501 (row-level security violation)", err)
			}
		})
	}
	if n := count(t, f.owner, "platform.rls_probe WHERE tenant_id = '"+f.b.String()+"'"); n != 1 {
		t.Fatalf("tenant B has %d probe rows after rejected writes, want 1", n)
	}
}

func TestWithTenantRejectsTheNilTenant(t *testing.T) {
	f := setup(t)
	called := false
	err := db.WithTenant(context.Background(), f.app, uuid.Nil, func(pgx.Tx) error { called = true; return nil })
	if !errors.Is(err, db.ErrNoTenant) {
		t.Fatalf("err = %v, want ErrNoTenant", err)
	}
	if called {
		t.Fatal("callback ran without a tenant")
	}
}

func TestTenantDoesNotLeakToTheNextUseOfAConnection(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	cfg := f.app.Config().Copy()
	cfg.MaxConns = 1 // force reuse of the same connection
	p, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	if err := db.WithTenant(ctx, p, f.a, func(pgx.Tx) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if n := count(t, p, "platform.tenant_modules"); n != 0 {
		t.Fatalf("after WithTenant committed, a plain query sees %d rows, want 0", n)
	}
	errBoom := errors.New("boom")
	if err := db.WithTenant(ctx, p, f.a, func(pgx.Tx) error { return errBoom }); !errors.Is(err, errBoom) {
		t.Fatalf("err = %v, want the callback error", err)
	}
	if n := count(t, p, "platform.tenant_modules"); n != 0 {
		t.Fatalf("after WithTenant rolled back, a plain query sees %d rows, want 0", n)
	}
}

func TestAppRoleCannotBypassRLS(t *testing.T) {
	f := setup(t)
	var super, bypass bool
	err := f.app.QueryRow(context.Background(),
		`SELECT rolsuper, rolbypassrls FROM pg_roles WHERE rolname = current_user`).Scan(&super, &bypass)
	if err != nil {
		t.Fatal(err)
	}
	if super || bypass {
		t.Fatalf("app role superuser=%v bypassrls=%v, want both false", super, bypass)
	}
	var owned int
	if err := f.app.QueryRow(context.Background(),
		`SELECT count(*) FROM pg_tables WHERE tableowner = current_user`).Scan(&owned); err != nil {
		t.Fatal(err)
	}
	if owned != 0 {
		t.Fatalf("app role owns %d tables, want 0", owned)
	}
}

// TestEveryTenantTableIsProtected guards future migrations: any table with a
// tenant_id column must have RLS enabled and forced.
func TestEveryTenantTableIsProtected(t *testing.T) {
	f := setup(t)
	rows, err := f.owner.Query(context.Background(), `
		SELECT c.oid::regclass::text
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		JOIN pg_attribute a ON a.attrelid = c.oid AND a.attname = 'tenant_id' AND NOT a.attisdropped
		WHERE c.relkind IN ('r', 'p')
		  AND n.nspname NOT IN ('pg_catalog', 'information_schema')
		  AND NOT (c.relrowsecurity AND c.relforcerowsecurity)`)
	if err != nil {
		t.Fatal(err)
	}
	unprotected, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		t.Fatal(err)
	}
	if len(unprotected) > 0 {
		t.Fatalf("tables with tenant_id but without ENABLE+FORCE RLS: %v", unprotected)
	}
}
