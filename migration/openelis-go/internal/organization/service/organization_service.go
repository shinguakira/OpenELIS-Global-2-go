// Package service ports org.openelisglobal.organization.service (the read
// paths backing the ported endpoints). Folder layout mirrors the Java
// source during migration.
package service

import (
	"fmt"
	"time"

	"openelis-go/internal/organization/daoimpl"
	"openelis-go/internal/organization/valueholder"
)

// OrganizationService holds no DB handle — all access goes through
// OrganizationDAOImpl (DAO-only-imports-ORM rule).
type OrganizationService struct {
	DAO *daoimpl.OrganizationDAOImpl
}

// GetAllTypes mirrors OrganizationTypeServiceImpl.getAll() (inherited,
// unoverridden, from BaseObjectServiceImpl) — GET rest/organization/types.
func (s *OrganizationService) GetAllTypes() ([]valueholder.OrganizationType, error) {
	return s.DAO.GetAllTypes()
}

// GetAll mirrors OrganizationServiceImpl.getAllOrganizations() — GET
// rest/organization-list. Populates each row's linked OrganizationTypes via
// one batched query (not N+1). testSections is not populated: confirmed via
// exploration (full-repo grep for setTestSections) that Organization.testSections
// is never written anywhere in the Java codebase — it always serializes as
// [] there, so the Go DTO layer hardcodes [] rather than modeling a field
// that carries no real data.
func (s *OrganizationService) GetAll() ([]valueholder.Organization, map[int64][]valueholder.OrganizationType, error) {
	orgs, err := s.DAO.GetAll()
	if err != nil {
		return nil, nil, err
	}
	ids := make([]int64, len(orgs))
	for i, o := range orgs {
		ids[i] = o.ID
	}
	types, err := s.DAO.GetTypesForOrgIDs(ids)
	if err != nil {
		return nil, nil, err
	}
	return orgs, types, nil
}

// GetByID mirrors OrganizationServiceImpl.get(id) (inherited BaseObjectServiceImpl.get)
// — GET rest/organization/{id}. Returns (nil, nil, nil) when not found — see
// OrganizationDAOImpl.GetByID's doc comment for why this port returns 404
// instead of reproducing Java's 500-on-not-found bug.
func (s *OrganizationService) GetByID(id int64) (*valueholder.Organization, []valueholder.OrganizationType, error) {
	org, err := s.DAO.GetByID(id)
	if err != nil || org == nil {
		return org, nil, err
	}
	typesByOrg, err := s.DAO.GetTypesForOrgIDs([]int64{id})
	if err != nil {
		return nil, nil, err
	}
	return org, typesByOrg[id], nil
}

// GetActiveChildrenByParentID mirrors DisplayListController.getDepartmentsForReferingSite:
// DAO returns all children (active + inactive), active-only filtering
// happens here (Java does this filter in the controller; kept in the
// service layer here per this port's own layering rule — same output,
// different layer). IsActive is nullable in the DB (confirmed — no NOT
// NULL/DEFAULT) even though Java's equivalent filter
// (org.getIsActive().equals(YES)) has no null guard, a latent NPE risk in
// Java; this port checks the pointer explicitly instead of inheriting that.
func (s *OrganizationService) GetActiveChildrenByParentID(parentID int64) ([]valueholder.Organization, error) {
	children, err := s.DAO.GetChildrenByParentID(parentID)
	if err != nil {
		return nil, err
	}
	active := make([]valueholder.Organization, 0, len(children))
	for _, c := range children {
		if c.IsActive != nil && *c.IsActive == "Y" {
			active = append(active, c)
		}
	}
	return active, nil
}

// GenerateSiteCode mirrors OrganizationServiceImpl.generateSiteCode():
// "S" + local-server-date (yyMMdd) + "-" + 5-digit zero-padded sequence
// value, e.g. "S260813-00042". Uses clinlims.site_code_seq, not the
// organization table's own organization_seq.
func (s *OrganizationService) GenerateSiteCode() (string, error) {
	seq, err := s.DAO.NextSiteCodeSeq()
	if err != nil {
		return "", err
	}
	date := time.Now().Format("060102") // yyMMdd
	return fmt.Sprintf("S%s-%05d", date, seq), nil
}
