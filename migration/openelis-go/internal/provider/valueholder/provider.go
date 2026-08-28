// Package valueholder ports org.openelisglobal.provider.valueholder and
// org.openelisglobal.person.valueholder (the Person fields Provider joins
// to). Folder layout mirrors the Java source during migration.
//
// Provider/Person are Hibernate-XML-mapped in Java (hibernate/hbm/Provider.hbm.xml,
// Person.hbm.xml), not JPA-annotated — column names/nullability here come
// from the real DB schema (db/dbInit/OpenELIS-Global.sql + liquibase
// 2.6.x.x/provider.xml, 2.8.x.x/desynchronous_objects.xml,
// 3.5.x.x/021-add-person-gps-columns.xml), not the hbm.xml metadata (which
// is stale in places — e.g. Provider.hbm.xml declares external_id length=10,
// the real column is varchar(50); Liquibase, not Hibernate, owns the schema).
package valueholder

import "time"

// Person mirrors person.valueholder.Person. Maps to clinlims.person. No DB
// column here has NOT NULL besides id — every other field is a pointer so a
// NULL scans as nil and is dropped from JSON (omitempty), matching Jackson's
// Include.NON_NULL. `patients` (a lazy, @JsonIgnore'd Set<Patient> in Java)
// is not ported — never touched or serialized by any of the ported
// endpoints.
type Person struct {
	ID            int64      `gorm:"column:id"`
	LastName      *string    `gorm:"column:last_name"`
	FirstName     *string    `gorm:"column:first_name"`
	MiddleName    *string    `gorm:"column:middle_name"`
	MultipleUnit  *string    `gorm:"column:multiple_unit"`
	StreetAddress *string    `gorm:"column:street_address"`
	City          *string    `gorm:"column:city"`
	State         *string    `gorm:"column:state"`
	ZipCode       *string    `gorm:"column:zip_code"`
	Country       *string    `gorm:"column:country"`
	WorkPhone     *string    `gorm:"column:work_phone"`
	HomePhone     *string    `gorm:"column:home_phone"`
	CellPhone     *string    `gorm:"column:cell_phone"`
	PrimaryPhone  *string    `gorm:"column:primary_phone"`
	Fax           *string    `gorm:"column:fax"`
	Email         *string    `gorm:"column:email"`
	GpsLatitude   *float64   `gorm:"column:gps_latitude"`
	GpsLongitude  *float64   `gorm:"column:gps_longitude"`
	Lastupdated   *time.Time `gorm:"column:lastupdated"`
}

// TableName pins the GORM table name.
func (Person) TableName() string { return "clinlims.person" }

// Provider mirrors provider.valueholder.Provider. Maps to clinlims.provider.
// PersonID is NOT NULL at the DB level (prov_person_fk, MATCH FULL) even
// though the Hibernate <many-to-one> mapping doesn't declare not-null="true"
// — confirmed via exploration that the DDL, not the hbm.xml, is the real
// constraint. There is no ProviderType lookup table anywhere in the schema
// or codebase — provider_type is a bare, unvalidated 1-char code.
type Provider struct {
	ID       int64   `gorm:"column:id"`
	NPI      *string `gorm:"column:npi"`
	PersonID int64   `gorm:"column:person_id"`
	// ExternalID: real DB column is varchar(50); Provider.hbm.xml's
	// length="10" is stale (Liquibase owns the schema, not Hibernate here).
	ExternalID   *string    `gorm:"column:external_id"`
	ProviderType *string    `gorm:"column:provider_type"`
	Lastupdated  *time.Time `gorm:"column:lastupdated"`
	// Active is nullable — Provider.java's getActive() coalesces a null
	// DB value to false at the Java object level (`active == null ? false
	// : active`). This port keeps the DAO-level value nullable and lets the
	// DTO/service layer apply that same coalesce, rather than baking the
	// coalesce into the DAO scan.
	Active         *bool   `gorm:"column:active"`
	FhirUUID       *string `gorm:"column:fhir_uuid"`
	Desynchronized bool    `gorm:"column:desynchronized"` // NOT NULL DEFAULT false
}

// TableName pins the GORM table name.
func (Provider) TableName() string { return "clinlims.provider" }
