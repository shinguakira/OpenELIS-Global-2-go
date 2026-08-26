// Package service ports org.openelisglobal.provider.service +
// org.openelisglobal.person.service (the read paths backing the ported
// endpoints). Folder layout mirrors the Java source during migration.
//
// Per constitution.md Layer III — the "Data Compilation Rule": this layer
// eagerly fetches and compiles ALL data needed for the response and returns
// the complete DTO. The controller (Layer IV) only does request/response
// mapping — no DTO shaping there.
package service

import (
	"strconv"

	"openelis-go/internal/provider/daoimpl"
	"openelis-go/internal/provider/form"
	"openelis-go/internal/provider/valueholder"
)

// ProviderService holds no DB handle — all access goes through
// ProviderDAOImpl (DAO-only-imports-ORM rule).
type ProviderService struct {
	DAO *daoimpl.ProviderDAOImpl
}

// GetProviderByID mirrors GET Provider/raw/{id}: providerService.get(id),
// with Person eagerly loaded to match Provider.hbm.xml's <many-to-one
// lazy="false">. Returns (nil, nil) when the provider doesn't exist — see
// ProviderDAOImpl.GetProviderByID's doc comment for the 404-not-500
// divergence from Java.
func (s *ProviderService) GetProviderByID(id int64) (*form.ProviderDTO, error) {
	p, err := s.DAO.GetProviderByID(id)
	if err != nil || p == nil {
		return nil, err
	}
	person, err := s.DAO.GetPersonByID(p.PersonID)
	if err != nil {
		return nil, err
	}
	var pv valueholder.Person
	if person != nil {
		pv = *person
	}
	dto := providerToDTO(*p, pv)
	return &dto, nil
}

// GetPersonByID mirrors GET Provider/Person/{id}: personService.get(id).
func (s *ProviderService) GetPersonByID(id int64) (*form.PersonDTO, error) {
	person, err := s.DAO.GetPersonByID(id)
	if err != nil || person == nil {
		return nil, err
	}
	dto := personToDTO(*person)
	return &dto, nil
}

// GetPractitionerByPersonID mirrors GET rest/practitioner
// (DisplayListController.getProviderInformation): despite the query param
// being named "providerId", it's actually a Person id — confirmed via
// exploration (every frontend caller passes a personId; DisplayListService's
// provider dropdown is keyed by Person id, not Provider id). Looks up the
// Person, then the Provider linked to that person. Returns (nil, nil) when
// the person doesn't exist, rather than reproducing Java's confirmed
// NullPointerException-on-unknown-id (personService.getPersonById(id)
// returning null, then immediately dereferenced by
// providerService.getProviderByPerson(null) -> person.getId() -> NPE,
// uncaught, surfaces as HTTP 500 today).
func (s *ProviderService) GetPractitionerByPersonID(personID int64) (*form.ProviderDTO, error) {
	person, err := s.DAO.GetPersonByID(personID)
	if err != nil || person == nil {
		return nil, err
	}
	provider, err := s.DAO.GetProviderByPersonID(personID)
	if err != nil || provider == nil {
		return nil, err
	}
	dto := providerToDTO(*provider, *person)
	return &dto, nil
}

// Search mirrors ProviderRestController.searchProviders: phone takes
// priority over search, falling back to an unfiltered paged listing when
// neither is given. page is 1-indexed (matches Java's convention exactly).
//
// Paging semantics were re-derived from live Java responses (not source
// alone) after a review found this port diverging on every edge case:
//
//   - The echoed page/pageSize are the caller's RAW values, never the
//     internally-clamped ones. Java echoes the raw request param
//     (ProviderRestController.java:133-134 puts the bound @RequestParam
//     straight into the response). Building the envelope here rather than in
//     the controller had accidentally promoted an internal clamp into the
//     wire contract: ?page=0&pageSize=0 answered "page":1,"pageSize":20.
//   - pageSize is a genuine ROW CAP, including zero. Java trims with
//     providers.subList(0, pageSize) (ProviderRestController.java:87-89), so
//     pageSize=0 returns an empty list — live-confirmed:
//     ?pageSize=0 -> {"pageSize":0,"page":1,"totalCount":3,"providers":[]}.
//     Treating 0 as "unset, use 20" (the old behavior here) contradicted
//     that. The absent-param default lives in the controller, matching
//     Java's @RequestParam(defaultValue="20"), so it is NOT duplicated here.
//   - Negative pageSize clamps to 0 (empty page). Java throws
//     IndexOutOfBoundsException from subList and 500s; not worth porting.
//
// Deliberate divergence from Java, documented in
// migration/b2-org-provider-migration.md: Java's DAOs internally cap the
// fetched rows at a *server config value* (page.defaultPageSize), not the
// caller's own pageSize — so a request for pageSize=50 silently returns at
// most page.defaultPageSize+1 rows today, regardless of what's asked for.
// This port has no equivalent config concept to replicate that against, and
// respecting the caller's real pageSize is more correct REST behavior
// besides, so LIMIT here is the caller's pageSize directly, with no hidden
// server-side ceiling.
//
// Second deliberate divergence: page < 1. Java computes
// startRecNo = ((page-1)*pageSize)+1 with no floor, hands the negative to
// Hibernate's setFirstResult (ProviderDAOImpl.java:168) and returns HTTP 500
// — live-confirmed for ?page=0&pageSize=20 and ?page=-3. This port floors
// page at 1 and serves the first page instead of reproducing a crash. The
// echoed page still reports the caller's raw value, so the response stays
// honest about what was asked for.
func (s *ProviderService) Search(search, phone string, page, pageSize int) (form.SearchResultDTO, error) {
	// Raw values, captured before any clamping — these are what get echoed.
	echoPage, echoPageSize := page, pageSize

	if page < 1 {
		page = 1
	}
	if pageSize < 0 {
		pageSize = 0
	}
	offset := (page - 1) * pageSize

	var rows []daoimpl.ProviderSearchRow
	var total int64
	var err error

	switch {
	case phone != "":
		rows, err = s.DAO.SearchByPhone(phone, offset, pageSize)
		if err != nil {
			return form.SearchResultDTO{}, err
		}
		total, err = s.DAO.CountByPhone(phone)
	case search != "":
		rows, err = s.DAO.SearchByName(search, offset, pageSize)
		if err != nil {
			return form.SearchResultDTO{}, err
		}
		total, err = s.DAO.CountByName(search)
	default:
		rows, err = s.DAO.GetPage(offset, pageSize)
		if err != nil {
			return form.SearchResultDTO{}, err
		}
		total, err = s.DAO.Count()
	}
	if err != nil {
		return form.SearchResultDTO{}, err
	}

	dtos := make([]form.ProviderSearchResultDTO, len(rows))
	for i, row := range rows {
		dtos[i] = searchRowToDTO(row)
	}
	return form.SearchResultDTO{Providers: dtos, TotalCount: total, Page: echoPage, PageSize: echoPageSize}, nil
}

// --- DTO shaping (constitution.md Layer III — belongs here, not the controller) ---

func personToDTO(p valueholder.Person) form.PersonDTO {
	dto := form.PersonDTO{
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

// providerToDTO's "active" coalesce mirrors Provider.java's getActive(),
// which coalesces a null DB value to false — done here at the DTO boundary
// rather than baked into the DAO/valueholder layer, so the DAO's Active
// field stays a faithful nullable mirror of the DB column.
func providerToDTO(p valueholder.Provider, person valueholder.Person) form.ProviderDTO {
	active := p.Active != nil && *p.Active
	fhirStr := ""
	if p.FhirUUID != nil {
		fhirStr = *p.FhirUUID
	}
	dto := form.ProviderDTO{
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

// searchRowToDTO mirrors the per-row Map-building loop in
// ProviderRestController.searchProviders exactly, including the fullName
// "Last, First" construction and the primaryPhone-falls-back-to-workPhone
// rule.
//
// isActive is a deliberate divergence from Java, documented in
// migration/b2-org-provider-migration.md: Java's own code is
// `providerData.put("isActive", "Y".equals(provider.getActive()))` —
// getActive() returns a Boolean, never the string "Y", so that comparison is
// always false and the real Java endpoint's isActive is always false today
// regardless of the actual active flag. This port returns the real value
// instead of reproducing that bug.
func searchRowToDTO(row daoimpl.ProviderSearchRow) form.ProviderSearchResultDTO {
	dto := form.ProviderSearchResultDTO{
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
