// Package rest ports org.openelisglobal.unitofmeasure.controller.rest.UnitOfMeasureRestController
// — GET /rest/uom (read; POST/create remains in Java).
// Folder layout mirrors the Java source during migration.
package rest

import (
	"net/http"
	"strconv"
	"strings"

	"openelis-go/internal/common/web"
	"openelis-go/internal/unitofmeasure/service"
	"openelis-go/internal/unitofmeasure/valueholder"
)

// uomDTO mirrors the {id, value} shape the Java controller returns.
// Java uses IdValuePair serialized as {"id": "...", "value": "..."}.
type uomDTO struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

func toDTO(u valueholder.UnitOfMeasure) uomDTO {
	return uomDTO{ID: strconv.FormatInt(u.ID, 10), Value: u.Name}
}

func toDTOs(list []valueholder.UnitOfMeasure) []uomDTO {
	dtos := make([]uomDTO, len(list))
	for i, u := range list {
		dtos[i] = toDTO(u)
	}
	return dtos
}

// UnitOfMeasureRestController mirrors UnitOfMeasureRestController.
type UnitOfMeasureRestController struct {
	Service *service.UnitOfMeasureService
}

// Routes registers GET /rest/uom (with optional ?type= param).
// Mirrors UnitOfMeasureRestController.getUnitOfMeasuresByType().
func Routes(mux *http.ServeMux, ctrl *UnitOfMeasureRestController) {
	web.Register(mux, "GET", "rest/uom", func(w http.ResponseWriter, r *http.Request) {
		t := strings.TrimSpace(r.URL.Query().Get("type"))
		var list []valueholder.UnitOfMeasure
		var err error
		if t != "" {
			list, err = ctrl.Service.GetUnitOfMeasuresByType(t)
		} else {
			list, err = ctrl.Service.GetAll()
		}
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, toDTOs(list))
	})
}
