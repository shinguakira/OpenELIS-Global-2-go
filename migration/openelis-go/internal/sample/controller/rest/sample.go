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
	"strconv"

	"openelis-go/internal/common/web"
	"openelis-go/internal/sample/form"
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

// OrderAttachmentRestController mirrors OrderAttachmentRestController
// (class-level @RequestMapping("/rest/order")) for the read endpoint only.
// The POST upload, the soft delete and the download/view routes are writes or
// binary streams and are out of the c2 read scope.
type OrderAttachmentRestController struct {
	Service *service.SampleService
}

// OrderAttachmentRoutes registers GET rest/order/{accessionNumber}/attachments.
func OrderAttachmentRoutes(mux *http.ServeMux, ctrl *OrderAttachmentRestController) {
	web.Register(mux, "GET", "rest/order/{accessionNumber}/attachments", func(w http.ResponseWriter, r *http.Request) {
		dtos, err := ctrl.Service.GetOrderAttachments(r.PathValue("accessionNumber"))
		if err != nil {
			log.Printf("c2: order attachments failed: %v", err)
			web.WriteJSON(w, http.StatusInternalServerError, form.ErrorDTO{Error: "Failed to save attachment"})
			return
		}
		if dtos == nil {
			// 404 WITH A BODY: Map.of("error", "Order not found"). Not the
			// bodiless 404 all-by-accession returns for the same missing
			// accession — two endpoints in this wave, two not-found shapes.
			web.WriteJSON(w, http.StatusNotFound, form.ErrorDTO{Error: "Order not found"})
			return
		}
		web.WriteJSON(w, http.StatusOK, dtos)
	})
}

// UnassignedSampleRestController mirrors UnassignedSampleRestController
// (class-level @RequestMapping("/rest/unassigned-sample")) for its GET
// endpoints. The three PUT routes are writes and out of the c2 read scope.
type UnassignedSampleRestController struct {
	Service *service.UnassignedSampleService
}

// UnassignedSampleRoutes registers the five unassigned-sample GET endpoints.
func UnassignedSampleRoutes(mux *http.ServeMux, ctrl *UnassignedSampleRestController) {
	// GET rest/unassigned-sample — a BARE @GetMapping, so the path is exactly
	// this with NO trailing slash. Spring 6 dropped automatic trailing-slash
	// matching, so "/rest/unassigned-sample/" is a 404 there and must be here
	// too. Go's ServeMux would happily treat the two as one pattern, so the
	// distinction has to be deliberate: registering only the bare form leaves
	// the slashed form unmatched, which is what produces the 404.
	web.Register(mux, "GET", "rest/unassigned-sample", func(w http.ResponseWriter, r *http.Request) {
		dtos, err := ctrl.Service.GetUnassignedForDashboard()
		if err != nil {
			log.Printf("c2: unassigned-sample failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, dtos)
	})

	// GET rest/unassigned-sample/items
	web.Register(mux, "GET", "rest/unassigned-sample/items", func(w http.ResponseWriter, r *http.Request) {
		dtos, err := ctrl.Service.GetUnassignedItems("")
		if err != nil {
			log.Printf("c2: unassigned-sample/items failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, dtos)
	})

	// GET rest/unassigned-sample/items/search?accessionNumber=
	//
	// accessionNumber is a required @RequestParam with no default, so Spring
	// answers 400 when it is absent — unlike /items, which takes no params.
	web.Register(mux, "GET", "rest/unassigned-sample/items/search", func(w http.ResponseWriter, r *http.Request) {
		if !r.URL.Query().Has("accessionNumber") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		dtos, err := ctrl.Service.GetUnassignedItems(r.URL.Query().Get("accessionNumber"))
		if err != nil {
			log.Printf("c2: unassigned-sample/items/search failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, dtos)
	})

	// GET rest/unassigned-sample/by-facility/{facilityId}
	//
	// facilityId binds as Integer, so a non-numeric value fails Spring's
	// binding with a 400 before the handler runs — in contrast with c1's
	// patient endpoints, where a String-bound path variable against a varchar
	// column simply matches nothing and returns 200.
	web.Register(mux, "GET", "rest/unassigned-sample/by-facility/{facilityId}", func(w http.ResponseWriter, r *http.Request) {
		facilityID, err := strconv.ParseInt(r.PathValue("facilityId"), 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		dtos, err := ctrl.Service.GetUnassignedByFacility(facilityID)
		if err != nil {
			log.Printf("c2: unassigned-sample/by-facility failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, dtos)
	})

	// GET rest/unassigned-sample/count-by-facility/{facilityId}
	//
	// Returns {"count": n} — a one-key HashMap, never a bare number. The count
	// is a SUBSET of by-facility's length, not its equal: see
	// UnassignedSampleService.CountUnassignedByFacility.
	web.Register(mux, "GET", "rest/unassigned-sample/count-by-facility/{facilityId}", func(w http.ResponseWriter, r *http.Request) {
		facilityID, err := strconv.ParseInt(r.PathValue("facilityId"), 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		count, err := ctrl.Service.CountUnassignedByFacility(facilityID)
		if err != nil {
			log.Printf("c2: unassigned-sample/count-by-facility failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, form.CountDTO{Count: count})
	})
}

// OrderDashboardRestController mirrors OrderSearchRestController's /dashboard
// endpoint (class-level @RequestMapping("/rest/order")).
type OrderDashboardRestController struct {
	Service *service.OrderDashboardService
}

// OrderDashboardRoutes registers GET rest/order/dashboard.
func OrderDashboardRoutes(mux *http.ServeMux, ctrl *OrderDashboardRestController) {
	web.Register(mux, "GET", "rest/order/dashboard", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		// Spring's @RequestParam defaults: page=1, pageSize=100,
		// includeExternal=false. A non-numeric page or pageSize would fail
		// Spring's binding with a 400; atoiDefault keeps the default instead,
		// which is a divergence on malformed input only and is noted here
		// rather than emulated — the c2 spec exercises neither.
		dto, err := ctrl.Service.GetDashboard(service.DashboardQuery{
			Page:            atoiDefault(q.Get("page"), 1),
			PageSize:        atoiDefault(q.Get("pageSize"), 100),
			Search:          q.Get("search"),
			Status:          q.Get("status"),
			Priority:        q.Get("priority"),
			IncludeExternal: q.Get("includeExternal") == "true",
			StartDate:       q.Get("startDate"),
			EndDate:         q.Get("endDate"),
		})
		if err != nil {
			log.Printf("c2: order/dashboard failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, dto)
	})
}

func atoiDefault(v string, def int) int {
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
