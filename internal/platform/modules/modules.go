// Package modules mounts business modules behind per-tenant activation.
//
// A module lives in internal/modules/<name>. Its root package is its public
// interface; everything else goes under internal/modules/<name>/internal/,
// which the Go compiler forbids other modules from importing. Modules talk to
// each other through explicit events, not through each other's internals.
package modules

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Jaed69/yupana/internal/platform/db"
	"github.com/Jaed69/yupana/internal/platform/tenant"
)

// Module is what a business module exposes to the API process.
type Module interface {
	// Name is the key in platform.tenant_modules and the route prefix /m/<name>/.
	Name() string
	// Register adds the module's routes, relative to its prefix.
	Register(mux *http.ServeMux)
}

// Mount serves each module under /m/<name>/. A request without a resolvable
// tenant gets 401; a tenant without the module active gets 403. Handlers find
// the tenant with tenant.FromContext.
func Mount(mux *http.ServeMux, pool *pgxpool.Pool, res tenant.Resolver, mods ...Module) {
	for _, m := range mods {
		sub := http.NewServeMux()
		m.Register(sub)
		prefix := "/m/" + m.Name()
		mux.Handle(prefix+"/", http.StripPrefix(prefix, guard(pool, res, m.Name(), sub)))
	}
}

func guard(pool *pgxpool.Pool, res tenant.Resolver, module string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := res.Resolve(r)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		active, err := IsActive(r.Context(), pool, id, module)
		if err != nil {
			slog.ErrorContext(r.Context(), "module activation check failed", "module", module, "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if !active {
			http.Error(w, "module not active", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r.WithContext(tenant.WithID(r.Context(), id)))
	})
}

// IsActive reports whether module is switched on for the tenant.
func IsActive(ctx context.Context, pool *pgxpool.Pool, tenantID uuid.UUID, module string) (bool, error) {
	var active bool
	err := db.WithTenant(ctx, pool, tenantID, func(tx pgx.Tx) error {
		// The explicit tenant filter is the first layer; RLS is the second.
		return tx.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM platform.tenant_modules WHERE tenant_id = $1 AND module = $2)`,
			tenantID, module).Scan(&active)
	})
	return active, err
}
