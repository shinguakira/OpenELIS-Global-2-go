// Package rest ports org.openelisglobal.unitofmeasure.controller.rest.UnitOfMeasureRestController
// — GET /rest/uom (read; POST/create remains in Java).
// Folder layout mirrors the Java source during migration.
//
// Per constitution.md Layer IV: request/response mapping and service calls
// only. See internal/unitofmeasure/form (Layer V) and
// internal/unitofmeasure/service (Layer III) for the DTO type and its shaping.
package rest

import (
	"net/http"
	"strings"

	"openelis-go/internal/common/web"
	"openelis-go/internal/unitofmeasure/form"
	"openelis-go/internal/unitofmeasure/service"
)

// UnitOfMeasureRestController mirrors UnitOfMeasureRestController.
type UnitOfMeasureRestController struct {
	Service *service.UnitOfMeasureService
}

// Routes registers GET /rest/uom (with optional ?type= param).
// Mirrors UnitOfMeasureRestController.getUnitOfMeasuresByType().
func Routes(mux *http.ServeMux, ctrl *UnitOfMeasureRestController) {
	web.Register(mux, "GET", "rest/uom", func(w http.ResponseWriter, r *http.Request) {
		t := strings.TrimSpace(r.URL.Query().Get("type"))
		var dtos []form.UnitOfMeasureDTO
		var err error
		if t != "" {
			dtos, err = ctrl.Service.GetUnitOfMeasuresByType(t)
		} else {
			dtos, err = ctrl.Service.GetAll()
		}
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, dtos)
	})
}
