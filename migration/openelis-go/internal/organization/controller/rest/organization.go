// Package rest ports org.openelisglobal.organization.controller.rest
// (+ the one org.openelisglobal.common.rest.DisplayListController method
// that's in scope, departments-for-site). Folder layout mirrors the Java
// source during migration.
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
	"openelis-go/internal/organization/valueholder"
)

// orgTypeDTO mirrors OrganizationType's JSON shape. Pointer fields with
// omitempty mirror Jackson's Include.NON_NULL (config/AppConfig.java) —
// dropped from the response when null, not emitted as JSON null.
type orgTypeDTO struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Description    *string `json:"description,omitempty"`
	HierarchyLevel *int    `json:"hierarchyLevel,omitempty"`
	Lastupdated    *int64  `json:"lastupdated,omitempty"`
}

func typeToDTO(t valueholder.OrganizationType) orgTypeDTO {
	dto := orgTypeDTO{
		ID:             strconv.FormatInt(t.ID, 10),
		Name:           t.Name,
		Description:    t.Description,
		HierarchyLevel: t.HierarchyLevel,
	}
	if t.Lastupdated != nil {
		ms := t.Lastupdated.UnixMilli()
		dto.Lastupdated = &ms
	}
	return dto
}

// organizationDTO mirrors the raw Organization entity's JSON shape (both
// organization-list and organization/{id} serialize the same entity type in
// Java, so they share this DTO here too). The self-referencing parent
// ("organization" field in Java, guarded there against infinite recursion
// by handleSelfReferencingParentOrg) is deliberately not embedded — see
// migration/b2-org-provider-migration.md. datim_org_code/datim_org_name are
// real DB columns but are never Hibernate-mapped in Java, so no endpoint
// ever serializes them; this DTO doesn't either.
type organizationDTO struct {
	ID                 string       `json:"id"`
	OrganizationName   string       `json:"organizationName"`
	City               *string      `json:"city,omitempty"`
	ZipCode            *string      `json:"zipCode,omitempty"`
	MlsSentinelLabFlag string       `json:"mlsSentinelLabFlag"`
	ShortName          *string      `json:"shortName,omitempty"`
	MultipleUnit       *string      `json:"multipleUnit,omitempty"`
	StreetAddress      *string      `json:"streetAddress,omitempty"`
	State              *string      `json:"state,omitempty"`
	InternetAddress    *string      `json:"internetAddress,omitempty"`
	CliaNum            *string      `json:"cliaNum,omitempty"`
	PwsID              *string      `json:"pwsId,omitempty"`
	Lastupdated        *int64       `json:"lastupdated,omitempty"`
	MlsLabFlag         *string      `json:"mlsLabFlag,omitempty"`
	IsActive           *string      `json:"isActive,omitempty"`
	LocalAbbrev        *string      `json:"organizationLocalAbbreviation,omitempty"`
	Code               *string      `json:"code,omitempty"`
	FhirUUID           *string      `json:"fhirUuid,omitempty"`
	OrganizationTypes  []orgTypeDTO `json:"organizationTypes"`
	// TestSections is permanently [] — confirmed via exploration
	// (full-repo grep) that Organization.testSections is never populated
	// anywhere in the Java codebase. Hardcoded here rather than modeled as
	// a real relationship, matching that dead-field reality exactly.
	TestSections []struct{} `json:"testSections"`
}

func orgToDTO(o valueholder.Organization, types []valueholder.OrganizationType) organizationDTO {
	dto := organizationDTO{
		ID:                 strconv.FormatInt(o.ID, 10),
		OrganizationName:   o.OrganizationName,
		City:               o.City,
		ZipCode:            o.ZipCode,
		MlsSentinelLabFlag: o.MlsSentinelLabFlag,
		ShortName:          o.ShortName,
		MultipleUnit:       o.MultipleUnit,
		StreetAddress:      o.StreetAddress,
		State:              o.State,
		InternetAddress:    o.InternetAddress,
		CliaNum:            o.CliaNum,
		PwsID:              o.PwsID,
		MlsLabFlag:         o.MlsLabFlag,
		IsActive:           o.IsActive,
		LocalAbbrev:        o.LocalAbbrev,
		Code:               o.Code,
		FhirUUID:           o.FhirUUID,
		OrganizationTypes:  make([]orgTypeDTO, len(types)),
		TestSections:       []struct{}{},
	}
	for i, t := range types {
		dto.OrganizationTypes[i] = typeToDTO(t)
	}
	if o.Lastupdated != nil {
		ms := o.Lastupdated.UnixMilli()
		dto.Lastupdated = &ms
	}
	return dto
}

// idValuePairDTO mirrors common.util.IdValuePair — {"id":..., "value":...}.
type idValuePairDTO struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

// OrganizationRestController mirrors OrganizationRestController for the
// endpoints in scope.
type OrganizationRestController struct {
	Service *service.OrganizationService
}

// Routes registers the in-scope organization endpoints.
func Routes(mux *http.ServeMux, ctrl *OrganizationRestController) {
	// GET rest/organization/types — mirrors getOrganizationTypes().
	web.Register(mux, "GET", "rest/organization/types", func(w http.ResponseWriter, r *http.Request) {
		types, err := ctrl.Service.GetAllTypes()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		dtos := make([]orgTypeDTO, len(types))
		for i, t := range types {
			dtos[i] = typeToDTO(t)
		}
		web.WriteJSON(w, http.StatusOK, dtos)
	})

	// GET rest/organization-list — mirrors getAllOrganizations().
	web.Register(mux, "GET", "rest/organization-list", func(w http.ResponseWriter, r *http.Request) {
		orgs, typesByOrg, err := ctrl.Service.GetAll()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		dtos := make([]organizationDTO, len(orgs))
		for i, o := range orgs {
			dtos[i] = orgToDTO(o, typesByOrg[o.ID])
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
		org, types, err := ctrl.Service.GetByID(id)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if org == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		web.WriteJSON(w, http.StatusOK, orgToDTO(*org, types))
	})

	// GET rest/organization/generate-site-code — mirrors generateSiteCode().
	web.Register(mux, "GET", "rest/organization/generate-site-code", func(w http.ResponseWriter, r *http.Request) {
		code, err := ctrl.Service.GenerateSiteCode()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, map[string]string{"siteCode": code})
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
		children, err := ctrl.Service.GetActiveChildrenByParentID(parentID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		dtos := make([]idValuePairDTO, len(children))
		for i, c := range children {
			dtos[i] = idValuePairDTO{ID: strconv.FormatInt(c.ID, 10), Value: c.OrganizationName}
		}
		web.WriteJSON(w, http.StatusOK, dtos)
	})
}
