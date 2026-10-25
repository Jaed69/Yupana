package modules_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/Jaed69/yupana/internal/platform/modules"
	"github.com/Jaed69/yupana/internal/platform/tenant"
	"github.com/Jaed69/yupana/internal/platform/testdb"
)

// echoModule answers with the tenant it was called for, proving the scope reaches handlers.
type echoModule struct{ name string }

func (m echoModule) Name() string { return m.name }

func (m echoModule) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /whoami", func(w http.ResponseWriter, r *http.Request) {
		id, _ := tenant.FromContext(r.Context())
		io.WriteString(w, id.String())
	})
}

func TestModuleRoutesRequireAnActiveModule(t *testing.T) {
	d := testdb.New(t)
	ctx := context.Background()
	active, inactive := uuid.New(), uuid.New()
	if _, err := d.Owner.Exec(ctx, `INSERT INTO platform.tenants (id, name) VALUES ($1, 'active'), ($2, 'inactive')`, active, inactive); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Owner.Exec(ctx, `INSERT INTO platform.tenant_modules (tenant_id, module) VALUES ($1, 'echo'), ($2, 'other')`, active, inactive); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	modules.Mount(mux, d.App, tenant.HeaderResolver{}, echoModule{name: "echo"})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	tests := []struct {
		name     string
		path     string
		tenantID string
		want     int
		body     string
	}{
		{name: "active module answers in the tenant scope", path: "/m/echo/whoami", tenantID: active.String(), want: 200, body: active.String()},
		{name: "module not active for the tenant", path: "/m/echo/whoami", tenantID: inactive.String(), want: 403},
		{name: "unknown tenant", path: "/m/echo/whoami", tenantID: uuid.NewString(), want: 403},
		{name: "no tenant", path: "/m/echo/whoami", want: 401},
		{name: "malformed tenant", path: "/m/echo/whoami", tenantID: "not-a-uuid", want: 401},
		{name: "unknown module", path: "/m/nope/whoami", tenantID: active.String(), want: 404},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, srv.URL+tt.path, nil)
			if tt.tenantID != "" {
				req.Header.Set(tenant.Header, tt.tenantID)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if resp.StatusCode != tt.want {
				t.Fatalf("status = %d, want %d (body %q)", resp.StatusCode, tt.want, body)
			}
			if tt.body != "" && string(body) != tt.body {
				t.Fatalf("body = %q, want %q", body, tt.body)
			}
		})
	}
}
