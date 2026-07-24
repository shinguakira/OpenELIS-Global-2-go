// Command openelis is the minimal Go re-implementation of the OpenELIS backend.
// It serves ported REST endpoints beside the Java WAR; the nginx proxy routes
// individual paths here as each endpoint passes parity (strangler-fig).
package main

import (
	"log"
	"net/http"
	"os"

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

	// a2: rest/supportedlocales{,/active,/fallback} — DB-backed. The static routes
	// above always work; these register only if Postgres is reachable.
	if database, err := db.Open(); err != nil {
		log.Printf("WARN: DB unavailable (%v); supportedlocales routes disabled", err)
	} else {
		svc := &localizationservice.SupportedLocaleService{
			DAO: &localizationdao.SupportedLocaleDAO{DB: database},
		}
		localizationrest.Routes(mux, svc)

		// Type-B status-type reads (status_of_sample).
		if statusSvc, err := commonservices.NewStatusService(database); err != nil {
			log.Printf("WARN: status service init failed (%v)", err)
		} else {
			commonrest.StatusRoutes(mux, statusSvc)
		}
		log.Printf("DB-backed routes enabled (supportedlocales, status-types)")
	}

	srv := &http.Server{Addr: addr, Handler: mux}
	log.Printf("openelis-go listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
