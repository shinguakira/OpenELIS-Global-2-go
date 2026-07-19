// Package web holds shared HTTP plumbing (mirrors the role of
// org.openelisglobal.common). This is migration-time scaffolding; the idiomatic
// Go reorganization happens at the end of the migration.
package web

import (
	"encoding/json"
	"net/http"
)

// WriteJSON writes v as a JSON response — the analog of Spring returning
// ResponseEntity with produces=APPLICATION_JSON_VALUE.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// Register mounts h at both the full proxied path
// (/api/OpenELIS-Global/<restPath>) and the bare /<restPath>, so the nginx
// proxy can forward a single path here unchanged.
func Register(mux *http.ServeMux, method, restPath string, h http.HandlerFunc) {
	mux.HandleFunc(method+" /api/OpenELIS-Global/"+restPath, h)
	mux.HandleFunc(method+" /"+restPath, h)
}
