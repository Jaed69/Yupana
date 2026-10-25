// Package tenant carries the tenant of a request and resolves it.
package tenant

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// Header is read by HeaderResolver. Dev only: a client must never choose its tenant in prod.
const Header = "X-Yupana-Tenant"

// Resolver identifies the tenant of an incoming request.
type Resolver interface {
	Resolve(r *http.Request) (uuid.UUID, bool)
}

// HeaderResolver trusts the Header value. It exists for local development
// until authentication lands; never wire it outside YUPANA_ENV=dev.
type HeaderResolver struct{}

func (HeaderResolver) Resolve(r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(r.Header.Get(Header))
	if err != nil || id == uuid.Nil {
		return uuid.Nil, false
	}
	return id, true
}

// DenyResolver resolves no tenant, so every tenant-scoped route fails closed.
type DenyResolver struct{}

func (DenyResolver) Resolve(*http.Request) (uuid.UUID, bool) { return uuid.Nil, false }

type ctxKey struct{}

// WithID returns a context scoped to tenant id.
func WithID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

// FromContext returns the tenant set by WithID.
func FromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(ctxKey{}).(uuid.UUID)
	return id, ok && id != uuid.Nil
}
