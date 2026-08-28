// Package valueholder ports org.openelisglobal.organization.valueholder.
// Folder layout mirrors the Java source during migration.
//
// Organization/OrganizationType are Hibernate-XML-mapped in Java (hibernate/hbm/
// Organization.hbm.xml, OrganizationType.hbm.xml), not JPA-annotated — column
// names/nullability here come from the real DB schema (db/dbInit/OpenELIS-Global.sql
// + liquibase/2.3.x.x/uuid_columns.xml), not the (partly stale) hbm.xml metadata.
// Confirmed stale hbm.xml facts NOT carried into this port: organizationName's
// hbm length=40 (real column is varchar(80)); externalId-equivalent lengths
// generally (Liquibase, not Hibernate, owns the schema).
//
// datim_org_code/datim_org_name exist in the DB but are not mapped by Hibernate
// at all (grepped: no hbm <property> for either) — every Java endpoint ignores
// them, so this port does too (matching API behavior, not schema completeness).
package valueholder

import "time"

// Organization mirrors organization.valueholder.Organization. Maps to
// clinlims.organization. Nullable DB columns use pointer fields so a NULL
// scans as nil and (with `json:",omitempty"` in the DTO layer) is dropped
// from the response entirely — matching Jackson's
// mapper.setSerializationInclusion(Include.NON_NULL) (config/AppConfig.java),
// which omits null fields rather than emitting JSON null.
type Organization struct {
	ID                 int64      `gorm:"column:id"`
	OrganizationName   string     `gorm:"column:name"`
	City               *string    `gorm:"column:city"`
	ZipCode            *string    `gorm:"column:zip_code"`
	MlsSentinelLabFlag string     `gorm:"column:mls_sentinel_lab_flag"` // NOT NULL DEFAULT 'N'
	OrgMltOrgMltID     *int64     `gorm:"column:org_mlt_org_mlt_id"`
	ParentOrgID        *int64     `gorm:"column:org_id"`
	ShortName          *string    `gorm:"column:short_name"`
	MultipleUnit       *string    `gorm:"column:multiple_unit"`
	StreetAddress      *string    `gorm:"column:street_address"`
	State              *string    `gorm:"column:state"`
	InternetAddress    *string    `gorm:"column:internet_address"`
	CliaNum            *string    `gorm:"column:clia_num"`
	PwsID              *string    `gorm:"column:pws_id"`
	Lastupdated        *time.Time `gorm:"column:lastupdated"`
	MlsLabFlag         *string    `gorm:"column:mls_lab_flag"`
	// IsActive is nullable at the DB level (no NOT NULL, no DEFAULT) even
	// though Java code elsewhere treats it as always "Y"/"N" — confirmed
	// exploration finding: DisplayListController.getDepartmentsForReferingSite
	// does org.getIsActive().equals(YES) with no null guard, a latent Java
	// NPE risk. Kept nullable here rather than assumed non-null, precisely
	// so the Go port doesn't inherit that same NPE-shaped bug.
	IsActive    *string `gorm:"column:is_active"`
	LocalAbbrev *string `gorm:"column:local_abbrev"`
	Code        *string `gorm:"column:code"`
	FhirUUID    *string `gorm:"column:fhir_uuid"`
}

// TableName pins the GORM table name.
func (Organization) TableName() string { return "clinlims.organization" }

// OrganizationType mirrors organization.valueholder.OrganizationType. Maps to
// clinlims.organization_type. Note the Java field "name" maps to DB column
// "short_name" (confirmed via OrganizationType.hbm.xml) — not "name".
type OrganizationType struct {
	ID             int64      `gorm:"column:id"`
	Name           string     `gorm:"column:short_name"`
	Description    *string    `gorm:"column:description"`
	HierarchyLevel *int       `gorm:"column:hierarchy_level"`
	Lastupdated    *time.Time `gorm:"column:lastupdated"`
}

// TableName pins the GORM table name.
func (OrganizationType) TableName() string { return "clinlims.organization_type" }
