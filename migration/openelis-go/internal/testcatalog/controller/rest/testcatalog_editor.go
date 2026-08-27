// Package rest ports org.openelisglobal.testcatalog.controller.rest.TestCatalogEditorRestController
// — three read-only reference endpoints: /lab-units, /sample-types, /panels.
// Folder layout mirrors the Java source during migration.
//
// Per constitution.md Layer IV: request/response mapping and service calls
// only — no DTO shaping or cross-service orchestration here (that used to live
// in this file directly against three foreign services; moved into
// internal/testcatalog/service, Layer III, which is the correct place for it).
package rest

import (
	"net/http"

	authmw "openelis-go/internal/auth/middleware"
	"openelis-go/internal/common/web"
	"openelis-go/internal/testcatalog/service"
)

// TestCatalogEditorRestController mirrors TestCatalogEditorRestController.
type TestCatalogEditorRestController struct {
	Service *service.TestCatalogEditorService
}

// Routes registers /rest/test-catalog/lab-units, /sample-types, /panels.
//
// ADMIN-GATED: the Java class carries a CLASS-level
// @PreAuthorize("hasRole('ADMIN')"), so it covers all three endpoints. Verified
// live — a non-admin gets Java's unhandled-AccessDeniedException 500, not a
// 403; see authmw.RequireAdmin for why that shape is reproduced rather than
// corrected. The gate is applied per route here because that is where Go can
// express a class-level annotation.
func Routes(mux *http.ServeMux, ctrl *TestCatalogEditorRestController) {
	web.Register(mux, "GET", "rest/test-catalog/lab-units", authmw.RequireAdmin(
		func(w http.ResponseWriter, r *http.Request) {
			dtos, err := ctrl.Service.GetLabUnits()
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			web.WriteJSON(w, http.StatusOK, dtos)
		}))

	web.Register(mux, "GET", "rest/test-catalog/sample-types", authmw.RequireAdmin(
		func(w http.ResponseWriter, r *http.Request) {
			dtos, err := ctrl.Service.GetSampleTypes()
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			web.WriteJSON(w, http.StatusOK, dtos)
		}))

	web.Register(mux, "GET", "rest/test-catalog/panels", authmw.RequireAdmin(
		func(w http.ResponseWriter, r *http.Request) {
			dtos, err := ctrl.Service.GetPanels()
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			web.WriteJSON(w, http.StatusOK, dtos)
		}))
}
