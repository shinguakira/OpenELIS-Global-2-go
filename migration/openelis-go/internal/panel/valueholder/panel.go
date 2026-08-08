// Package valueholder ports org.openelisglobal.panel.valueholder.
// Folder layout mirrors the Java source during migration.
package valueholder

// Panel mirrors panel/valueholder/Panel.java — one row of clinlims.panel.
// ID is int64 — the real Postgres type; string conversion for JSON happens in
// the controller DTO.
type Panel struct {
	ID        int64  `gorm:"column:id"`
	PanelName string `gorm:"column:panel_name"`
}

// TableName pins the GORM table name.
func (Panel) TableName() string { return "clinlims.panel" }
