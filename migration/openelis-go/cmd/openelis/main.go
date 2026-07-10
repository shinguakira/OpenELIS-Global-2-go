// Command openelis is the minimal Go re-implementation of the OpenELIS backend.
// It serves ported REST endpoints beside the Java WAR; the nginx proxy routes
// individual paths here as each endpoint passes parity (strangler-fig).
package main

import (
	"log"
	"net/http"
	"os"

	"openelis-go/internal/rest"
)

func main() {
	addr := os.Getenv("OE_GO_ADDR")
	if addr == "" {
		addr = ":8090"
	}

	srv := &http.Server{Addr: addr, Handler: rest.NewRouter()}
	log.Printf("openelis-go listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
