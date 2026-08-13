// Package service ports org.openelisglobal.provider.service +
// org.openelisglobal.person.service (the read paths backing the ported
// endpoints). Folder layout mirrors the Java source during migration.
package service

import (
	"openelis-go/internal/provider/daoimpl"
	"openelis-go/internal/provider/valueholder"
)

// ProviderService holds no DB handle — all access goes through
// ProviderDAOImpl (DAO-only-imports-ORM rule).
type ProviderService struct {
	DAO *daoimpl.ProviderDAOImpl
}

// GetProviderByID mirrors GET Provider/raw/{id}: providerService.get(id),
// with Person eagerly loaded to match Provider.hbm.xml's <many-to-one
// lazy="false">. Returns (nil, nil, nil) when the provider doesn't exist —
// see ProviderDAOImpl.GetProviderByID's doc comment for the 404-not-500
// divergence from Java.
func (s *ProviderService) GetProviderByID(id int64) (*valueholder.Provider, *valueholder.Person, error) {
	p, err := s.DAO.GetProviderByID(id)
	if err != nil || p == nil {
		return nil, nil, err
	}
	person, err := s.DAO.GetPersonByID(p.PersonID)
	if err != nil {
		return nil, nil, err
	}
	return p, person, nil
}

// GetPersonByID mirrors GET Provider/Person/{id}: personService.get(id).
func (s *ProviderService) GetPersonByID(id int64) (*valueholder.Person, error) {
	return s.DAO.GetPersonByID(id)
}

// GetPractitionerByPersonID mirrors GET rest/practitioner
// (DisplayListController.getProviderInformation): despite the query param
// being named "providerId", it's actually a Person id — confirmed via
// exploration (every frontend caller passes a personId; DisplayListService's
// provider dropdown is keyed by Person id, not Provider id). Looks up the
// Person, then the Provider linked to that person. Returns (nil, nil, nil)
// when the person doesn't exist, rather than reproducing Java's confirmed
// NullPointerException-on-unknown-id (personService.getPersonById(id)
// returning null, then immediately dereferenced by
// providerService.getProviderByPerson(null) -> person.getId() -> NPE,
// uncaught, surfaces as HTTP 500 today).
func (s *ProviderService) GetPractitionerByPersonID(personID int64) (*valueholder.Provider, *valueholder.Person, error) {
	person, err := s.DAO.GetPersonByID(personID)
	if err != nil || person == nil {
		return nil, nil, err
	}
	provider, err := s.DAO.GetProviderByPersonID(personID)
	if err != nil || provider == nil {
		return nil, person, err
	}
	return provider, person, nil
}

// SearchResult is the (rows, totalCount) pair GET provider/search needs.
type SearchResult struct {
	Providers  []daoimpl.ProviderSearchRow
	TotalCount int64
}

// Search mirrors ProviderRestController.searchProviders: phone takes
// priority over search, falling back to an unfiltered paged listing when
// neither is given. page is 1-indexed (matches Java's convention exactly).
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
func (s *ProviderService) Search(search, phone string, page, pageSize int) (SearchResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var rows []daoimpl.ProviderSearchRow
	var total int64
	var err error

	switch {
	case phone != "":
		rows, err = s.DAO.SearchByPhone(phone, offset, pageSize)
		if err != nil {
			return SearchResult{}, err
		}
		total, err = s.DAO.CountByPhone(phone)
	case search != "":
		rows, err = s.DAO.SearchByName(search, offset, pageSize)
		if err != nil {
			return SearchResult{}, err
		}
		total, err = s.DAO.CountByName(search)
	default:
		rows, err = s.DAO.GetPage(offset, pageSize)
		if err != nil {
			return SearchResult{}, err
		}
		total, err = s.DAO.Count()
	}
	if err != nil {
		return SearchResult{}, err
	}
	return SearchResult{Providers: rows, TotalCount: total}, nil
}
