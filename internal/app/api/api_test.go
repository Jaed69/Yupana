package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/Jaed69/yupana/internal/platform/config"
	"github.com/Jaed69/yupana/internal/platform/tenant"
)

// In prod no request can pick its tenant through the dev header: the route
// fails closed before touching the database.
func TestProdIgnoresTheDevTenantHeader(t *testing.T) {
	h := newHandler(config.Config{Env: "prod"}, nil)
	req := httptest.NewRequest(http.MethodGet, "/m/ingest/status", nil)
	req.Header.Set(tenant.Header, uuid.NewString())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestHealthz(t *testing.T) {
	rec := httptest.NewRecorder()
	newHandler(config.Config{Env: "prod"}, nil).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}
