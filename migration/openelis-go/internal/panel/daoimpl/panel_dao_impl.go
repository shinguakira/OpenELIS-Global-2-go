// Package daoimpl ports org.openelisglobal.panel.daoimpl.
// Folder layout mirrors the Java source during migration.
package daoimpl

import (
	"gorm.io/gorm"

	"openelis-go/internal/panel/valueholder"
)

// PanelDAOImpl ports PanelDAOImpl — reads clinlims.panel via GORM's query builder.
type PanelDAOImpl struct {
	DB *gorm.DB
}

// GetAllActivePanels mirrors PanelDAOImpl.getAllActivePanels():
// active panels (is_active = 'Y') ordered by name.
func (d *PanelDAOImpl) GetAllActivePanels() ([]valueholder.Panel, error) {
	var panels []valueholder.Panel
	result := d.DB.
		Select("id, COALESCE(name, '') AS panel_name").
		Where("is_active = ?", "Y").
		Order("name").
		Find(&panels)
	return panels, result.Error
}
