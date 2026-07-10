package rest

import (
	"encoding/json"
	"net/http"
)

// NewRouter wires the HTTP endpoints. Ported endpoints get registered here,
// one at a time, as each passes parity. For now it only exposes a health probe.
func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	return mux
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "UP"})
}

// writeJSON encodes v as a JSON response with the given status.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
