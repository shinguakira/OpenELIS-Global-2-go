// Package rest ports org.openelisglobal.sample.controller.rest.SampleRestController
// (class-level @RequestMapping("/rest/sample")) for the c2 read endpoints.
// Folder layout mirrors the Java source during migration.
//
// Per constitution.md Layer IV: request/response mapping and service calls
// only — the DTO lives in internal/sample/form (Layer V) and the decisions in
// internal/sample/service (Layer III).
//
// AUTH: both routes go through web.Register, which is default-deny since P0,
// so an unauthenticated caller gets Java's 302 to /LoginPage. Neither path has
// a system_module_url row and SampleRestController carries no @PreAuthorize,
// so authentication is the only gate — checked, not assumed.
package rest

import (
	"log"
	"net/http"

	"openelis-go/internal/common/web"
	"openelis-go/internal/sample/service"
)

// SampleRestController groups the c2 sample reads.
type SampleRestController struct {
	Service *service.SampleService
}

// Routes registers the c2 sample endpoints.
func Routes(mux *http.ServeMux, ctrl *SampleRestController) {
	// GET rest/sample/all-by-accession/{accessionNumber}
	//
	// Java's status split, reproduced:
	//   no such sample                    -> 404 (bodiless)
	//   sample with no NotStarted rows    -> 200 []
	//   otherwise                         -> 200 [rows]
	//
	// The controller wraps everything in `catch (Exception)` -> 500, so a
	// query failure is a 500 rather than a propagated error page. Same here.
	web.Register(mux, "GET", "rest/sample/all-by-accession/{accessionNumber}", func(w http.ResponseWriter, r *http.Request) {
		forms, err := ctrl.Service.GetAllByAccession(r.PathValue("accessionNumber"))
		if err != nil {
			log.Printf("c2: all-by-accession failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if forms == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		web.WriteJSON(w, http.StatusOK, forms)
	})

	// GET rest/sample/unassigned-by-accession/{accessionNumber}
	//
	// ALWAYS 500 — see SampleService.GetUnassignedByAccession. Java's HQL is
	// invalid and throws at parse time, so no input can succeed and the
	// controller's own not-found branch is unreachable. Registered rather than
	// omitted so a client gets the same status Java gives it; pinned, not
	// fixed.
	web.Register(mux, "GET", "rest/sample/unassigned-by-accession/{accessionNumber}", func(w http.ResponseWriter, r *http.Request) {
		if _, err := ctrl.Service.GetUnassignedByAccession(r.PathValue("accessionNumber")); err != nil {
			// Bodiless, matching ResponseEntity.status(500).build().
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Unreachable, exactly as in Java.
		w.WriteHeader(http.StatusInternalServerError)
	})
}

// PendingAnalysisRestController mirrors
// PendingAnalysisForTestProviderRestController (class-level
// @RequestMapping("/rest")).
type PendingAnalysisRestController struct {
	Service *service.PendingAnalysisService
}

// PendingAnalysisRoutes registers GET rest/getPendingAnalysisForTestProvider.
func PendingAnalysisRoutes(mux *http.ServeMux, ctrl *PendingAnalysisRestController) {
	web.Register(mux, "GET", "rest/getPendingAnalysisForTestProvider", func(w http.ResponseWriter, r *http.Request) {
		testID := r.URL.Query().Get("testId")

		// testId is @RequestParam WITHOUT required=false, so Spring rejects a
		// MISSING param at binding with its own 400 before the handler runs.
		// A PRESENT-but-blank one reaches the handler and hits
		// GenericValidator.isBlankOrNull, which returns 400 with this exact
		// plain-text body. Two different 400s; only the second carries a body.
		if testID == "" {
			if !r.URL.Query().Has("testId") {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			http.Error(w, "Internal error, please contact Admin and file bug report", http.StatusBadRequest)
			return
		}

		dto, err := ctrl.Service.GetPendingForTest(testID)
		if err != nil {
			log.Printf("c2: getPendingAnalysisForTestProvider failed: %v", err)
			http.Error(w, "Internal error, please contact Admin and file bug report", http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, dto)
	})
}
