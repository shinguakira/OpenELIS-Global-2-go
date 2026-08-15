// Package service ports org.openelisglobal.organization.service (the read
// paths backing the ported endpoints). Folder layout mirrors the Java
// source during migration.
//
// Per constitution.md Layer III — the "Data Compilation Rule": this layer
// eagerly fetches and compiles ALL data needed for the response and returns
// the complete DTO. The controller (Layer IV) only does request/response
// mapping — no DTO shaping there.
package service

import (
	"fmt"
	"strconv"
	"time"

	"openelis-go/internal/organization/daoimpl"
	"openelis-go/internal/organization/form"
	"openelis-go/internal/organization/valueholder"
)

// OrganizationService holds no DB handle — all access goes through
// OrganizationDAOImpl (DAO-only-imports-ORM rule).
type OrganizationService struct {
	DAO *daoimpl.OrganizationDAOImpl
}

// GetAllTypes mirrors OrganizationTypeServiceImpl.getAll() (inherited,
// unoverridden, from BaseObjectServiceImpl) — GET rest/organization/types.
func (s *OrganizationService) GetAllTypes() ([]form.OrgTypeDTO, error) {
	types, err := s.DAO.GetAllTypes()
	if err != nil {
		return nil, err
	}
	dtos := make([]form.OrgTypeDTO, len(types))
	for i, t := range types {
		dtos[i] = typeToDTO(t)
	}
	return dtos, nil
}

// GetAll mirrors OrganizationServiceImpl.getAllOrganizations() — GET
// rest/organization-list. Populates each row's linked OrganizationTypes via
// one batched query (not N+1). testSections is not populated: confirmed via
// exploration (full-repo grep for setTestSections) that Organization.testSections
// is never written anywhere in the Java codebase — it always serializes as
// [] there, so the DTO layer hardcodes [] rather than modeling a field that
// carries no real data.
func (s *OrganizationService) GetAll() ([]form.OrganizationDTO, error) {
	orgs, err := s.DAO.GetAll()
	if err != nil {
		return nil, err
	}
	ids := make([]int64, len(orgs))
	for i, o := range orgs {
		ids[i] = o.ID
	}
	typesByOrg, err := s.DAO.GetTypesForOrgIDs(ids)
	if err != nil {
		return nil, err
	}
	dtos := make([]form.OrganizationDTO, len(orgs))
	for i, o := range orgs {
		dtos[i] = orgToDTO(o, typesByOrg[o.ID])
	}
	return dtos, nil
}

// GetByID mirrors OrganizationServiceImpl.get(id) (inherited BaseObjectServiceImpl.get)
// — GET rest/organization/{id}. Returns (nil, nil) when not found — see
// OrganizationDAOImpl.GetByID's doc comment for why this port returns 404
// instead of reproducing Java's 500-on-not-found bug.
func (s *OrganizationService) GetByID(id int64) (*form.OrganizationDTO, error) {
	org, err := s.DAO.GetByID(id)
	if err != nil || org == nil {
		return nil, err
	}
	typesByOrg, err := s.DAO.GetTypesForOrgIDs([]int64{id})
	if err != nil {
		return nil, err
	}
	dto := orgToDTO(*org, typesByOrg[id])
	return &dto, nil
}

// GetActiveChildrenByParentID mirrors DisplayListController.getDepartmentsForReferingSite:
// DAO returns all children (active + inactive), active-only filtering
// happens here (Java does this filter in the controller; kept in the
// service layer here per this port's own layering rule — same output,
// different layer). IsActive is nullable in the DB (confirmed — no NOT
// NULL/DEFAULT) even though Java's equivalent filter
// (org.getIsActive().equals(YES)) has no null guard, a latent NPE risk in
// Java; this port checks the pointer explicitly instead of inheriting that.
func (s *OrganizationService) GetActiveChildrenByParentID(parentID int64) ([]form.IdValuePairDTO, error) {
	children, err := s.DAO.GetChildrenByParentID(parentID)
	if err != nil {
		return nil, err
	}
	dtos := make([]form.IdValuePairDTO, 0, len(children))
	for _, c := range children {
		if c.IsActive != nil && *c.IsActive == "Y" {
			dtos = append(dtos, form.IdValuePairDTO{ID: strconv.FormatInt(c.ID, 10), Value: c.OrganizationName})
		}
	}
	return dtos, nil
}

// GenerateSiteCode mirrors OrganizationServiceImpl.generateSiteCode():
// "S" + server-date (yyMMdd) + "-" + 5-digit zero-padded sequence value,
// e.g. "S260813-00042". Uses clinlims.site_code_seq, not the organization
// table's own organization_seq.
//
// Explicitly UTC, not host-local time. Java's LocalDate.now() resolves via
// ZoneId.systemDefault() — the JVM's system timezone — and docker-compose.yml
// pins that container to TZ=${TZ:-UTC} (confirmed live: `docker exec
// openelisglobal-webapp date` reports UTC). Using Go's local time.Now()
// instead of UTC was caught as a real divergence during live Java-vs-Go
// comparison: run on a non-UTC host (this dev machine is JST), the two
// servers produced different site-code dates for requests made in the same
// instant (JST 00:00-09:00 is still "yesterday" in UTC). Pinning to UTC here
// makes Go's output match Java's actual configured behavior regardless of
// what host/container the Go binary itself happens to run on.
func (s *OrganizationService) GenerateSiteCode() (form.SiteCodeDTO, error) {
	seq, err := s.DAO.NextSiteCodeSeq()
	if err != nil {
		return form.SiteCodeDTO{}, err
	}
	date := time.Now().UTC().Format("060102") // yyMMdd, UTC — see doc comment above
	return form.SiteCodeDTO{SiteCode: fmt.Sprintf("S%s-%05d", date, seq)}, nil
}

// --- DTO shaping (constitution.md Layer III — belongs here, not the controller) ---

func typeToDTO(t valueholder.OrganizationType) form.OrgTypeDTO {
	dto := form.OrgTypeDTO{
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

func orgToDTO(o valueholder.Organization, types []valueholder.OrganizationType) form.OrganizationDTO {
	dto := form.OrganizationDTO{
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
		OrganizationTypes:  make([]form.OrgTypeDTO, len(types)),
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
