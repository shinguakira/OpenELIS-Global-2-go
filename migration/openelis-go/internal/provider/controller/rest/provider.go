// Package rest ports org.openelisglobal.provider.controller.rest (+ the one
// org.openelisglobal.common.rest.DisplayListController method that's in
// scope, practitioner). Folder layout mirrors the Java source during
// migration.
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
	"openelis-go/internal/provider/daoimpl"
	"openelis-go/internal/provider/service"
	"openelis-go/internal/provider/valueholder"
)

// personDTO mirrors Person's JSON shape (GET Provider/Person/{id}). Pointer
// fields with omitempty mirror Jackson's Include.NON_NULL.
type personDTO struct {
	ID            string   `json:"id"`
	LastName      *string  `json:"lastName,omitempty"`
	FirstName     *string  `json:"firstName,omitempty"`
	MiddleName    *string  `json:"middleName,omitempty"`
	MultipleUnit  *string  `json:"multipleUnit,omitempty"`
	StreetAddress *string  `json:"streetAddress,omitempty"`
	City          *string  `json:"city,omitempty"`
	State         *string  `json:"state,omitempty"`
	ZipCode       *string  `json:"zipCode,omitempty"`
	Country       *string  `json:"country,omitempty"`
	WorkPhone     *string  `json:"workPhone,omitempty"`
	HomePhone     *string  `json:"homePhone,omitempty"`
	CellPhone     *string  `json:"cellPhone,omitempty"`
	PrimaryPhone  *string  `json:"primaryPhone,omitempty"`
	Fax           *string  `json:"fax,omitempty"`
	Email         *string  `json:"email,omitempty"`
	GpsLatitude   *float64 `json:"gpsLatitude,omitempty"`
	GpsLongitude  *float64 `json:"gpsLongitude,omitempty"`
	// Lastupdated: confirmed missing via live Java-vs-Go comparison — Java
	// serializes this on Person (and on Provider, below) as epoch millis;
	// the valueholder already scans the column, this DTO just wasn't
	// surfacing it. Same encoding as organizationDTO.Lastupdated.
	Lastupdated *int64 `json:"lastupdated,omitempty"`
}

func personToDTO(p valueholder.Person) personDTO {
	dto := personDTO{
		ID:            strconv.FormatInt(p.ID, 10),
		LastName:      p.LastName,
		FirstName:     p.FirstName,
		MiddleName:    p.MiddleName,
		MultipleUnit:  p.MultipleUnit,
		StreetAddress: p.StreetAddress,
		City:          p.City,
		State:         p.State,
		ZipCode:       p.ZipCode,
		Country:       p.Country,
		WorkPhone:     p.WorkPhone,
		HomePhone:     p.HomePhone,
		CellPhone:     p.CellPhone,
		PrimaryPhone:  p.PrimaryPhone,
		Fax:           p.Fax,
		Email:         p.Email,
		GpsLatitude:   p.GpsLatitude,
		GpsLongitude:  p.GpsLongitude,
	}
	if p.Lastupdated != nil {
		ms := p.Lastupdated.UnixMilli()
		dto.Lastupdated = &ms
	}
	return dto
}

// providerDTO mirrors Provider's JSON shape (GET Provider/raw/{id}, GET
// rest/practitioner) — the full entity with Person nested under "person",
// matching Provider.hbm.xml's eager (lazy="false") many-to-one.
// fhirUuidAsString is a real second field on Provider.java
// (getFhirUuidAsString(), a Jackson-visible getter distinct from the UUID
// getFhirUuid()) — both are ported since Java emits both.
type providerDTO struct {
	ID               string    `json:"id"`
	ExternalID       *string   `json:"externalId,omitempty"`
	NPI              *string   `json:"npi,omitempty"`
	ProviderType     *string   `json:"providerType,omitempty"`
	Person           personDTO `json:"person"`
	FhirUUID         *string   `json:"fhirUuid,omitempty"`
	FhirUUIDAsString string    `json:"fhirUuidAsString"`
	Active           bool      `json:"active"`
	Desynchronized   bool      `json:"desynchronized"`
	// Lastupdated: same confirmed-missing gap as personDTO.Lastupdated —
	// see that field's doc comment.
	Lastupdated *int64 `json:"lastupdated,omitempty"`
}

func providerToDTO(p valueholder.Provider, person valueholder.Person) providerDTO {
	// Provider.java's getActive() coalesces a null DB value to false —
	// ported here at the DTO boundary rather than baking the coalesce into
	// the DAO/valueholder layer, so the DAO's Active field stays a
	// faithful nullable mirror of the DB column.
	active := p.Active != nil && *p.Active
	fhirStr := ""
	if p.FhirUUID != nil {
		fhirStr = *p.FhirUUID
	}
	dto := providerDTO{
		ID:               strconv.FormatInt(p.ID, 10),
		ExternalID:       p.ExternalID,
		NPI:              p.NPI,
		ProviderType:     p.ProviderType,
		Person:           personToDTO(person),
		FhirUUID:         p.FhirUUID,
		FhirUUIDAsString: fhirStr,
		Active:           active,
		Desynchronized:   p.Desynchronized,
	}
	if p.Lastupdated != nil {
		ms := p.Lastupdated.UnixMilli()
		dto.Lastupdated = &ms
	}
	return dto
}

// providerSearchResultDTO mirrors the hand-built Map<String,Object> row
// shape from ProviderRestController.searchProviders exactly (field-by-field,
// not a formal Java DTO class — see migration exploration notes).
type providerSearchResultDTO struct {
	ID         string  `json:"id"`
	PersonID   *string `json:"personId,omitempty"`
	FirstName  *string `json:"firstName,omitempty"`
	LastName   *string `json:"lastName,omitempty"`
	Name       *string `json:"name,omitempty"`
	Phone      *string `json:"phone,omitempty"`
	Fax        *string `json:"fax,omitempty"`
	Email      *string `json:"email,omitempty"`
	ExternalID *string `json:"externalId,omitempty"`
	IsActive   bool    `json:"isActive"`
}

// searchRowToDTO mirrors the per-row Map-building loop in
// ProviderRestController.searchProviders exactly, including the
// fullName "Last, First" construction and the primaryPhone-falls-back-to-
// workPhone rule.
//
// isActive is a deliberate divergence from Java, documented in
// migration/b2-org-provider-migration.md: Java's own code is
// `providerData.put("isActive", "Y".equals(provider.getActive()))` —
// getActive() returns a Boolean, never the string "Y", so that comparison
// is always false and the real Java endpoint's isActive is always false
// today regardless of the actual active flag. This port returns the real
// value instead of reproducing that bug.
func searchRowToDTO(row daoimpl.ProviderSearchRow) providerSearchResultDTO {
	dto := providerSearchResultDTO{
		ID:         strconv.FormatInt(row.ProviderID, 10),
		ExternalID: row.ExternalID,
		IsActive:   row.Active != nil && *row.Active,
	}
	personIDStr := strconv.FormatInt(row.PersonID, 10)
	dto.PersonID = &personIDStr
	dto.FirstName = row.FirstName
	dto.LastName = row.LastName

	fullName := ""
	if row.LastName != nil {
		fullName = *row.LastName
	}
	if row.FirstName != nil {
		if fullName != "" {
			fullName += ", "
		}
		fullName += *row.FirstName
	}
	dto.Name = &fullName

	phone := row.PrimaryPhone
	if phone == nil || *phone == "" {
		phone = row.WorkPhone
	}
	dto.Phone = phone
	dto.Fax = row.Fax
	dto.Email = row.Email
	return dto
}

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
		provider, person, err := ctrl.Service.GetProviderByID(id)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if provider == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		var p valueholder.Person
		if person != nil {
			p = *person
		}
		web.WriteJSON(w, http.StatusOK, providerToDTO(*provider, p))
	})

	// GET rest/Provider/Person/{id} — mirrors getPerson(id). Real 404 on
	// not-found (same divergence as above).
	web.Register(mux, "GET", "rest/Provider/Person/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		person, err := ctrl.Service.GetPersonByID(id)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if person == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		web.WriteJSON(w, http.StatusOK, personToDTO(*person))
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
		provider, person, err := ctrl.Service.GetPractitionerByPersonID(id)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if provider == nil || person == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		web.WriteJSON(w, http.StatusOK, providerToDTO(*provider, *person))
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

		result, err := ctrl.Service.Search(q.Get("search"), q.Get("phone"), page, pageSize)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		dtos := make([]providerSearchResultDTO, len(result.Providers))
		for i, row := range result.Providers {
			dtos[i] = searchRowToDTO(row)
		}
		web.WriteJSON(w, http.StatusOK, map[string]any{
			"providers":  dtos,
			"totalCount": result.TotalCount,
			"page":       page,
			"pageSize":   pageSize,
		})
	})
}
