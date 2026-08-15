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

	"openelis-go/internal/common/web"
	"openelis-go/internal/testcatalog/service"
)

// TestCatalogEditorRestController mirrors TestCatalogEditorRestController.
type TestCatalogEditorRestController struct {
	Service *service.TestCatalogEditorService
}

// Routes registers /rest/test-catalog/lab-units, /sample-types, /panels.
func Routes(mux *http.ServeMux, ctrl *TestCatalogEditorRestController) {
	web.Register(mux, "GET", "rest/test-catalog/lab-units", func(w http.ResponseWriter, r *http.Request) {
		dtos, err := ctrl.Service.GetLabUnits()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, dtos)
	})

	web.Register(mux, "GET", "rest/test-catalog/sample-types", func(w http.ResponseWriter, r *http.Request) {
		dtos, err := ctrl.Service.GetSampleTypes()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, dtos)
	})

	web.Register(mux, "GET", "rest/test-catalog/panels", func(w http.ResponseWriter, r *http.Request) {
		dtos, err := ctrl.Service.GetPanels()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, dtos)
	})
}
