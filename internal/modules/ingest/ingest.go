// Package ingest is the public interface of the invoice ingestion module.
// Phase 0 only mounts it; capture and parsing arrive in the ingestion phase.
package ingest

import (
	"encoding/json"
	"net/http"
)

// Module implements modules.Module.
type Module struct{}

func (Module) Name() string { return "ingest" }

func (Module) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /status", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"module": "ingest", "status": "ok"})
	})
}
