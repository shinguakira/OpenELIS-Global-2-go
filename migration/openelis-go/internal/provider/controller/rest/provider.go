// Package rest ports org.openelisglobal.provider.controller.rest (+ the one
// org.openelisglobal.common.rest.DisplayListController method that's in
// scope, practitioner). Folder layout mirrors the Java source during
// migration.
//
// Per constitution.md Layer IV: request/response mapping and service calls
// only — no DTO shaping here. See internal/provider/form (Layer V) and
// internal/provider/service (Layer III) for the DTO types and how they're
// built.
//
// Endpoints deliberately NOT ported in this pass — see
// migration/b2-org-provider-migration.md for the full writeup:
//   - GET rest/ProviderMenu, rest/SearchProviderMenu: Struts-legacy
//     AdminMenuForm envelope (same pattern/complexity as OrganizationMenu),
//     no e2e contract exists yet, and the Java controller double-fetches
//     the same query per request (real Java-side waste, not a semantic
//     requirement worth porting faithfully).
package rest

import (
	"net/http"
	"strconv"

	"openelis-go/internal/common/web"
	"openelis-go/internal/provider/service"
)

// ProviderRestController mirrors ProviderRestController + the
// DisplayListController.practitioner method, for the endpoints in scope.
type ProviderRestController struct {
	Service *service.ProviderService
}

// Routes registers the in-scope provider endpoints.
func Routes(mux *http.ServeMux, ctrl *ProviderRestController) {
	// GET rest/Provider/raw/{id} — mirrors getProvider(id). ProviderRestController
	// carries a class-level @RequestMapping("/rest") in Java (confirmed by
	// reading the source directly after a live-Java comparison 404'd on the
	// un-prefixed path) — every route in this file needs the same rest/
	// prefix. Real 404 on not-found; see ProviderDAOImpl.GetProviderByID's
	// doc comment for why this diverges from Java's confirmed
	// 500-on-not-found bug.
	web.Register(mux, "GET", "rest/Provider/raw/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		dto, err := ctrl.Service.GetProviderByID(id)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if dto == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		web.WriteJSON(w, http.StatusOK, dto)
	})

	// GET rest/Provider/Person/{id} — mirrors getPerson(id). Real 404 on
	// not-found (same divergence as above).
	web.Register(mux, "GET", "rest/Provider/Person/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		dto, err := ctrl.Service.GetPersonByID(id)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if dto == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		web.WriteJSON(w, http.StatusOK, dto)
	})

	// GET rest/practitioner?providerId=<personId> — mirrors
	// DisplayListController.getProviderInformation. Param name is
	// misleading (it's actually a Person id) but kept exactly as-is to
	// match every real caller. Real 404 when the person doesn't exist
	// (Java NPEs uncaught there today — see ProviderService.GetPractitionerByPersonID's
	// doc comment) or when no provider is linked to that person.
	web.Register(mux, "GET", "rest/practitioner", func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Query().Get("providerId")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "providerId is required", http.StatusBadRequest)
			return
		}
		dto, err := ctrl.Service.GetPractitionerByPersonID(id)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if dto == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		web.WriteJSON(w, http.StatusOK, dto)
	})

	// GET rest/provider/search?search=&phone=&page=&pageSize= — mirrors
	// ProviderRestController.searchProviders.
	web.Register(mux, "GET", "rest/provider/search", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		page := 1
		if v := q.Get("page"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				page = n
			}
		}
		pageSize := 20
		if v := q.Get("pageSize"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				pageSize = n
			}
		}

		dto, err := ctrl.Service.Search(q.Get("search"), q.Get("phone"), page, pageSize)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, dto)
	})
}
