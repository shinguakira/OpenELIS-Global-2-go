// Package rest ports org.openelisglobal.testconfiguration.controller.rest.TestCatalogRestController
// — GET /rest/TestCatalog (read-only).
// Folder layout mirrors the Java source during migration.
//
// Per constitution.md Layer IV: request/response mapping and service calls
// only. See internal/testconfiguration/form (Layer V) and
// internal/testconfiguration/service (Layer III) for the DTO types and how
// they're built.
package rest

import (
	"net/http"

	authmw "openelis-go/internal/auth/middleware"
	"openelis-go/internal/common/web"
	"openelis-go/internal/testconfiguration/service"
)

// TestCatalogRestController mirrors TestCatalogRestController.
type TestCatalogRestController struct {
	Service *service.TestCatalogService
}

// Routes registers GET /rest/TestCatalog.
//
// This endpoint is gated TWICE in Java, by two independent mechanisms, and it
// is the only ported route where that is true:
//
//  1. ModuleAuthenticationInterceptor — `/TestCatalog` is one of the 382 rows in
//     system_module_url, mapped to the `TestCatalog` module (held by the Global
//     Administrator and Test Management roles). Applied globally by
//     auth/middleware.Guard, not here.
//  2. @PreAuthorize("hasRole('ADMIN')") at class level — applied here.
//
// The order is observable, and all three outcomes are verified live:
//
//	e2e_reception (no module, not admin)   -> 401 { "status": 401, ... }
//	e2e_testmgmt  (has module, not admin)  -> 500 (unhandled AccessDeniedException)
//	admin                                  -> 200
//
// Until P0 auth landed this route was open to any caller, which is exactly the
// "Go is more permissive than Java and nothing flags it" case
// auth-adoption-plan.md §9.2 predicted.
func Routes(mux *http.ServeMux, ctrl *TestCatalogRestController) {
	web.Register(mux, "GET", "rest/TestCatalog", authmw.RequireAdmin(
		func(w http.ResponseWriter, r *http.Request) {
			dto, err := ctrl.Service.BuildForm()
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			web.WriteJSON(w, http.StatusOK, dto)
		}))
}
