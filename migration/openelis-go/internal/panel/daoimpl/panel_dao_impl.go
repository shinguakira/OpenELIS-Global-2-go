// Package daoimpl ports org.openelisglobal.panel.daoimpl.
// Folder layout mirrors the Java source during migration.
package daoimpl

import (
	"database/sql"

	"openelis-go/internal/panel/valueholder"
)

// PanelDAOImpl ports PanelDAOImpl — reads clinlims.panel.
type PanelDAOImpl struct {
	DB *sql.DB
}

// GetAllActivePanels mirrors PanelDAOImpl.getAllActivePanels():
// active panels (is_active = 'Y') ordered by name.
func (d *PanelDAOImpl) GetAllActivePanels() ([]valueholder.Panel, error) {
	rows, err := d.DB.Query(`
		SELECT id::text, COALESCE(name, '') AS panel_name
		FROM clinlims.panel
		WHERE is_active = 'Y'
		ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []valueholder.Panel{}
	for rows.Next() {
		var p valueholder.Panel
		if err := rows.Scan(&p.ID, &p.PanelName); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, rows.Err()
}
