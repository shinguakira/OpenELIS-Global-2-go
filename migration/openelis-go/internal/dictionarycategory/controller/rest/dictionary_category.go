// Package rest ports org.openelisglobal.dictionary.controller.rest.DictionaryMenuRestController
// — GET /rest/dictionary-categories only (read; writes remain in Java).
// Folder layout mirrors the Java source during migration.
//
// Per constitution.md Layer IV (Controllers): request/response mapping and
// service calls only — no DTO shaping here. See internal/dictionarycategory/form
// (Layer V) for the DTO type and internal/dictionarycategory/service (Layer III)
// for how it's built.
package rest

import (
	"net/http"

	authmw "openelis-go/internal/auth/middleware"
	"openelis-go/internal/common/web"
	"openelis-go/internal/dictionarycategory/service"
)

// DictionaryMenuRestController mirrors DictionaryMenuRestController.
type DictionaryMenuRestController struct {
	Service *service.DictionaryCategoryService
}

// Routes registers GET /rest/dictionary-categories.
// Mirrors DictionaryMenuRestController.fetchDictionaryCategories().
//
// ADMIN-GATED: DictionaryMenuRestController carries a CLASS-level
// @PreAuthorize("hasRole('ADMIN')"), so this read is admin-only in Java even
// though it looks like ordinary reference data. Verified live: a non-admin gets
// a 500, not a 403 — see authmw.RequireAdmin. The path has no system_module_url
// row, so the module check auto-allows it and this is the only gate.
func Routes(mux *http.ServeMux, ctrl *DictionaryMenuRestController) {
	web.Register(mux, "GET", "rest/dictionary-categories", authmw.RequireAdmin(
		func(w http.ResponseWriter, r *http.Request) {
			dtos, err := ctrl.Service.GetAll()
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			web.WriteJSON(w, http.StatusOK, dtos)
		}))
}
