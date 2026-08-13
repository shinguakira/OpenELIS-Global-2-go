// Package daoimpl ports org.openelisglobal.organization.daoimpl (+ the parts
// of org.openelisglobal.common.daoimpl.BaseDAOImpl actually exercised by the
// ported endpoints). Folder layout mirrors the Java source during migration.
package daoimpl

import (
	"gorm.io/gorm"

	"openelis-go/internal/organization/valueholder"
)

// OrganizationDAOImpl ports OrganizationDAOImpl + the generic BaseDAOImpl
// paths it actually uses for the ported (read-only) endpoints.
type OrganizationDAOImpl struct {
	DB *gorm.DB
}

// GetAll mirrors OrganizationDAOImpl.getAllOrganizations() — HQL "from
// Organization": no WHERE (active AND inactive), no ORDER BY (DB-natural
// order). Used by GET rest/organization-list.
func (d *OrganizationDAOImpl) GetAll() ([]valueholder.Organization, error) {
	var orgs []valueholder.Organization
	result := d.DB.Find(&orgs)
	if orgs == nil {
		orgs = []valueholder.Organization{}
	}
	return orgs, result.Error
}

// GetByID mirrors the generic BaseObjectServiceImpl.get(id) -> BaseDAOImpl.get(id)
// path used by GET rest/organization/{id}: entityManager.find(Organization.class, id),
// a plain SELECT * FROM organization WHERE id = ?. Returns (nil, nil) on no
// match — unlike Java, which throws ObjectNotFoundException there (a
// confirmed bug: the controller's `if (organization != null)` branch is dead
// code, so any not-found id returns HTTP 500, not 404, on the real Java app
// today). This port returns a real 404 instead of reproducing that — see
// migration/b2-org-provider-migration.md for the divergence writeup.
func (d *OrganizationDAOImpl) GetByID(id int64) (*valueholder.Organization, error) {
	var org valueholder.Organization
	result := d.DB.First(&org, "id = ?", id)
	if result.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &org, nil
}

// GetChildrenByParentID mirrors OrganizationDAOImpl.getOrganizationsByParentId:
// HQL "from Organization o where o.organization.id = :parentId order by o.id"
// — children of parentId, active AND inactive (active-only filtering happens
// in the service layer here, matching where Java's controller does it:
// DisplayListController.getDepartmentsForReferingSite).
func (d *OrganizationDAOImpl) GetChildrenByParentID(parentID int64) ([]valueholder.Organization, error) {
	var orgs []valueholder.Organization
	result := d.DB.Where("org_id = ?", parentID).Order("id ASC").Find(&orgs)
	if orgs == nil {
		orgs = []valueholder.Organization{}
	}
	return orgs, result.Error
}

// GetAllTypes mirrors the code path GET rest/organization/types actually
// uses: OrganizationTypeServiceImpl doesn't override getAll(), so it
// resolves to the generic BaseObjectServiceImpl.getAll() ->
// getAllOrdered(["id"], false) -> a CriteriaQuery with no WHERE and
// ORDER BY id ASC. (OrganizationTypeDAOImpl also defines its own
// getAllOrganizationTypes() with HQL "from OrganizationType" and no ORDER
// BY, but that method is never called by this endpoint — confirmed via
// trace, don't port its no-sort behavior by mistake.)
func (d *OrganizationDAOImpl) GetAllTypes() ([]valueholder.OrganizationType, error) {
	var types []valueholder.OrganizationType
	result := d.DB.Order("id ASC").Find(&types)
	if types == nil {
		types = []valueholder.OrganizationType{}
	}
	return types, result.Error
}

// orgTypeLink is the scan target for the organization_organization_type join
// table joined to organization_type, used to batch-load each organization's
// linked types in one query instead of N+1.
type orgTypeLink struct {
	OrgID          int64
	ID             int64
	Name           string
	Description    *string
	HierarchyLevel *int
}

// GetTypesForOrgIDs batch-loads the OrganizationType rows linked to each of
// the given organization ids, via the organization_organization_type M:N
// join table. Mirrors what Organization.organizationTypes (a lazy,
// inverse="true" Hibernate <set>) resolves to when populated. Eagerly
// loading here (rather than trying to replicate Hibernate's lazy-init
// timing, which the exploration found to be environment-dependent) is a
// deliberate, documented choice — see migration/b2-org-provider-migration.md.
// Returns a map keyed by organization id; ids with no linked types are
// simply absent from the map (caller treats that as an empty slice).
func (d *OrganizationDAOImpl) GetTypesForOrgIDs(orgIDs []int64) (map[int64][]valueholder.OrganizationType, error) {
	result := map[int64][]valueholder.OrganizationType{}
	if len(orgIDs) == 0 {
		return result, nil
	}
	var links []orgTypeLink
	err := d.DB.Table("clinlims.organization_organization_type AS link").
		Select("link.org_id AS org_id, t.id AS id, t.short_name AS name, t.description AS description, t.hierarchy_level AS hierarchy_level").
		Joins("JOIN clinlims.organization_type t ON t.id = link.org_type_id").
		Where("link.org_id IN ?", orgIDs).
		Order("link.org_id ASC, t.id ASC").
		Find(&links).Error
	if err != nil {
		return nil, err
	}
	for _, l := range links {
		result[l.OrgID] = append(result[l.OrgID], valueholder.OrganizationType{
			ID:             l.ID,
			Name:           l.Name,
			Description:    l.Description,
			HierarchyLevel: l.HierarchyLevel,
		})
	}
	return result, nil
}

// NextSiteCodeSeq mirrors OrganizationServiceImpl.generateSiteCode()'s
// native query: SELECT nextval('clinlims.site_code_seq'). A bare sequence
// nextval isn't a table read — no valueholder involved — so Raw() here is
// the same class of legitimate exception the plan docs describe (the
// Hibernate-native-SQL-equivalent tier), not the Raw()-for-everything
// anti-pattern the a2/b1 rewrites eliminated.
func (d *OrganizationDAOImpl) NextSiteCodeSeq() (int64, error) {
	var seq int64
	err := d.DB.Raw("SELECT nextval('clinlims.site_code_seq')").Scan(&seq).Error
	return seq, err
}
