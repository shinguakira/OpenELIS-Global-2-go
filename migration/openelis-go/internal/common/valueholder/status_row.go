// Package valueholder holds DB projections for internal/common domains.
package valueholder

// StatusRow is the raw projection of one clinlims.status_of_sample row.
// ID is int64 — the real Postgres type; StatusService converts to string
// (strconv.FormatInt) when building its in-memory id/label map, matching
// Java's BaseObject<String> id contract at the point it's actually needed.
type StatusRow struct {
	ID         int64  `gorm:"column:id"`
	StatusType string `gorm:"column:status_type"`
	Name       string `gorm:"column:name"`
	DisplayKey string `gorm:"column:display_key"`
}

// TableName pins the GORM table name.
func (StatusRow) TableName() string { return "clinlims.status_of_sample" }
