// Package rest ports org.openelisglobal.organization.controller.rest
// (+ the one org.openelisglobal.common.rest.DisplayListController method
// that's in scope, departments-for-site). Folder layout mirrors the Java
// source during migration.
//
// Per constitution.md Layer IV: request/response mapping and service calls
// only — no DTO shaping here. See internal/organization/form (Layer V) and
// internal/organization/service (Layer III) for the DTO types and how
// they're built.
//
// Endpoints deliberately NOT ported in this pass — see
// migration/b2-org-provider-migration.md for the full writeup:
//   - GET rest/Organization: Type-D form-load (embeds departmentList,
//     orgTypes, selectedTypes, address-part scalars) — not a Type-B
//     reference read, much larger scope than this pass.
//   - GET/rest/OrganizationMenu, rest/SearchOrganizationMenu: real Java bug
//     (totalRecordCount always shows the grand total, even on a filtered
//     search) plus a Struts-legacy AdminMenuForm envelope; no e2e contract
//     exists yet to pin the correct shape.
//   - GET rest/OrganizationExport: needs the FHIR transform layer
//     (FhirTransformService) — out of scope for a Type-B read pass; also
//     the migration plan's own D5 principle is facade-not-reimplement for
//     FHIR.
package rest

import (
	"net/http"
	"strconv"

	"openelis-go/internal/common/web"
	"openelis-go/internal/organization/service"
)

// OrganizationRestController mirrors OrganizationRestController for the
// endpoints in scope.
type OrganizationRestController struct {
	Service *service.OrganizationService
}

// Routes registers the in-scope organization endpoints.
func Routes(mux *http.ServeMux, ctrl *OrganizationRestController) {
	// GET rest/organization/types — mirrors getOrganizationTypes().
	web.Register(mux, "GET", "rest/organization/types", func(w http.ResponseWriter, r *http.Request) {
		dtos, err := ctrl.Service.GetAllTypes()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, dtos)
	})

	// GET rest/organization-list — mirrors getAllOrganizations().
	web.Register(mux, "GET", "rest/organization-list", func(w http.ResponseWriter, r *http.Request) {
		dtos, err := ctrl.Service.GetAll()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, dtos)
	})

	// GET rest/organization/{id} — mirrors getOrganization(id). Real 404 on
	// not-found; see OrganizationDAOImpl.GetByID's doc comment for why this
	// diverges from Java's confirmed 500-on-not-found bug. Registered as an
	// exact {id} pattern (Go 1.22+ ServeMux), not a "/rest/organization/"
	// subtree catch-all — so it can never shadow the more specific
	// rest/organization/types or rest/organization/generate-site-code
	// routes regardless of registration order.
	web.Register(mux, "GET", "rest/organization/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		dto, err := ctrl.Service.GetByID(id)
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

	// GET rest/organization/generate-site-code — mirrors generateSiteCode().
	web.Register(mux, "GET", "rest/organization/generate-site-code", func(w http.ResponseWriter, r *http.Request) {
		dto, err := ctrl.Service.GenerateSiteCode()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, dto)
	})

	// GET rest/departments-for-site?refferingSiteId=... — mirrors
	// DisplayListController.getDepartmentsForReferingSite. Query param name
	// (misspelled) matches the real Java param and every frontend caller
	// exactly — not a typo introduced here.
	web.Register(mux, "GET", "rest/departments-for-site", func(w http.ResponseWriter, r *http.Request) {
		parentIDStr := r.URL.Query().Get("refferingSiteId")
		if parentIDStr == "" {
			http.Error(w, "refferingSiteId is required", http.StatusBadRequest)
			return
		}
		parentID, err := strconv.ParseInt(parentIDStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid refferingSiteId", http.StatusBadRequest)
			return
		}
		dtos, err := ctrl.Service.GetActiveChildrenByParentID(parentID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, dtos)
	})
}
