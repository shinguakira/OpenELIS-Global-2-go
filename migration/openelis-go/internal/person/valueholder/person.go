// Package valueholder ports org.openelisglobal.person.valueholder.
// Folder layout mirrors the Java source during migration.
//
// NOTE for whoever merges b2 (migration/b2-org-provider): that branch defines
// an equivalent Person struct under internal/provider/valueholder because the
// provider endpoints needed it first. Java's real home for this type is
// org.openelisglobal.person.valueholder, which is where this one lives. When
// the two branches meet, keep THIS one and repoint provider at it rather than
// carrying two definitions of the same table.
package valueholder

import "time"

// Person mirrors person.valueholder.Person. Maps to clinlims.person.
// Every column except id is nullable in the real schema (verified against
// information_schema), so all of them are pointer fields: a NULL scans as nil
// and, with `json:",omitempty"` at the DTO boundary, is dropped from the
// response entirely — matching Jackson's Include.NON_NULL rather than
// emitting a JSON null.
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
	Fax           *string    `gorm:"column:fax"`
	Email         *string    `gorm:"column:email"`
	PrimaryPhone  *string    `gorm:"column:primary_phone"`
	GpsLatitude   *float64   `gorm:"column:gps_latitude"`
	GpsLongitude  *float64   `gorm:"column:gps_longitude"`
	Lastupdated   *time.Time `gorm:"column:lastupdated"`
}

// TableName pins the GORM table name.
func (Person) TableName() string { return "clinlims.person" }
