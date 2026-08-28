// Package rest ports
// org.openelisglobal.genericsample.controller.rest.GenericSampleOrderRestController
// (class-level @RequestMapping("/rest")) for its READ endpoint. Folder layout
// mirrors the Java source during migration.
//
// AUTH: the route goes through web.Register, which is default-deny since P0, so
// an unauthenticated caller gets Java's 302 to /LoginPage. The path has no
// system_module_url row and the controller carries no @PreAuthorize, so
// authentication is the only gate — checked, not assumed.
package rest

import (
	"net/http"

	"openelis-go/internal/common/web"
	"openelis-go/internal/genericsample/form"
	"openelis-go/internal/genericsample/service"
)

// GenericSampleOrderRestController groups the Wave 4.5 read.
type GenericSampleOrderRestController struct {
	Service *service.GenericSampleOrderService
}

// Routes registers GET rest/GenericSampleOrder.
//
// Three outcomes, all measured against live Java:
//
//	no accessionNumber param      -> 400 with Spring's ProblemDetail envelope
//	accession with no sample      -> 404 {"error":"No sample found with accession number: X"}
//	accession that EXISTS         -> 500 {"error":"Failed to retrieve generic sample order: …"}
//
// The last one is a JAVA DEFECT, reproduced rather than fixed — see
// service.ErrGenericSampleOrderRollback. It gives this endpoint an inverted
// success contract: the only input that answers cleanly is one that matches
// nothing.
func Routes(mux *http.ServeMux, ctrl *GenericSampleOrderRestController) {
	web.Register(mux, "GET", "rest/GenericSampleOrder", func(w http.ResponseWriter, r *http.Request) {
		if !r.URL.Query().Has("accessionNumber") {
			// @RequestParam(required = true) with no defaultValue. Spring
			// rejects this at binding, before the handler runs, and answers its
			// own ProblemDetail — five keys, NOT the {error} shape the handler
			// itself produces. Two different 400-vs-404 envelopes on one
			// endpoint, so they are built separately.
			web.WriteJSON(w, http.StatusBadRequest, form.MissingAccessionNumberProblem(r))
			return
		}

		accessionNumber := r.URL.Query().Get("accessionNumber")
		exists, err := ctrl.Service.SampleExists(accessionNumber)
		if err != nil {
			web.WriteJSON(w, http.StatusInternalServerError, form.ErrorDTO{
				Error: "Failed to retrieve generic sample order: " + err.Error(),
			})
			return
		}
		if !exists {
			web.WriteJSON(w, http.StatusNotFound, form.ErrorDTO{
				Error: "No sample found with accession number: " + accessionNumber,
			})
			return
		}
		// The sample exists, which in Java is precisely when it fails.
		web.WriteJSON(w, http.StatusInternalServerError, form.ErrorDTO{
			Error: service.GenericSampleOrderRollbackMessage,
		})
	})
}
