// Command openelis is the minimal Go re-implementation of the OpenELIS backend.
// It serves ported REST endpoints beside the Java WAR; the nginx proxy routes
// individual paths here as each endpoint passes parity (strangler-fig).
package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"openelis-go/internal/common/db"
	commonrest "openelis-go/internal/common/rest"
	commonservices "openelis-go/internal/common/services"
	"openelis-go/internal/common/web"
	localizationrest "openelis-go/internal/localization/controller/rest"
	localizationdao "openelis-go/internal/localization/daoimpl"
	localizationservice "openelis-go/internal/localization/service"
	systemrest "openelis-go/internal/system/controller/rest"
	calculatedrest "openelis-go/internal/testcalculated/controller/rest"
)

func main() {
	addr := os.Getenv("OE_GO_ADDR")
	if addr == "" {
		addr = ":8090"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		web.WriteJSON(w, http.StatusOK, map[string]string{"status": "UP"})
	})

	// Register each domain's REST routes (mirrors Spring auto-discovering
	// @RestController beans). One line per ported domain.
	systemrest.Routes(mux)     // a1: rest/server-time
	calculatedrest.Routes(mux) // a2: rest/math-functions
	commonrest.Routes(mux)     // a2: rest/sample-item-status-types

	// a2: rest/supportedlocales{,/active,/fallback} + status-type endpoints —
	// DB-backed. Retry connecting to Postgres so a slow Compose startup (where
	// the DB container starts before Postgres is ready to accept connections)
	// does not permanently disable these routes. On exhaustion we log.Fatal so
	// the restart policy (restart: unless-stopped) retries the whole process.
	const maxAttempts = 10
	const retryDelay = 3 * time.Second

	var database *sql.DB
	for i := 1; i <= maxAttempts; i++ {
		conn, err := db.Open()
		if err == nil {
			database = conn
			break
		}
		if i == maxAttempts {
			log.Fatalf("DB unavailable after %d attempts (%v); exiting so restart policy can retry", maxAttempts, err)
		}
		log.Printf("DB not ready (attempt %d/%d): %v — retrying in %s", i, maxAttempts, err, retryDelay)
		time.Sleep(retryDelay)
	}

	svc := &localizationservice.SupportedLocaleService{
		DAO: &localizationdao.SupportedLocaleDAO{DB: database},
	}
	localizationrest.Routes(mux, svc)

	if statusSvc, err := commonservices.NewStatusService(database); err != nil {
		log.Printf("WARN: status service init failed (%v)", err)
	} else {
		commonrest.StatusRoutes(mux, statusSvc)
	}
	log.Printf("DB-backed routes enabled (supportedlocales, status-types)")

	srv := &http.Server{Addr: addr, Handler: mux}
	log.Printf("openelis-go listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
