// Package rest ports org.openelisglobal.testconfiguration.controller.rest.TestCatalogRestController
// — GET /rest/TestCatalog (read-only; admin-gated in Java but open in Go during migration).
// Folder layout mirrors the Java source during migration.
//
// Per constitution.md Layer IV: request/response mapping and service calls
// only. See internal/testconfiguration/form (Layer V) and
// internal/testconfiguration/service (Layer III) for the DTO types and how
// they're built.
package rest

import (
	"net/http"

	"openelis-go/internal/common/web"
	"openelis-go/internal/testconfiguration/service"
)

// TestCatalogRestController mirrors TestCatalogRestController.
type TestCatalogRestController struct {
	Service *service.TestCatalogService
}

// Routes registers GET /rest/TestCatalog.
func Routes(mux *http.ServeMux, ctrl *TestCatalogRestController) {
	web.Register(mux, "GET", "rest/TestCatalog", func(w http.ResponseWriter, r *http.Request) {
		dto, err := ctrl.Service.BuildForm()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, dto)
	})
}
