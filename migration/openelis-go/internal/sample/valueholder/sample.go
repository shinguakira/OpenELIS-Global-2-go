// Package valueholder ports org.openelisglobal.sample.valueholder for the c2
// read paths. Folder layout mirrors the Java source during migration.
package valueholder

// Sample maps clinlims.sample. Only the columns the c2 reads need are declared;
// the table has many more, and pulling them in would invite emitting fields
// Java's DTOs never expose.
type Sample struct {
	ID              int64  `gorm:"column:id"`
	AccessionNumber string `gorm:"column:accession_number"`
	StatusID        *int64 `gorm:"column:status_id"`
}

// TableName pins the schema-qualified table (GORM would guess "samples").
func (Sample) TableName() string { return "clinlims.sample" }
