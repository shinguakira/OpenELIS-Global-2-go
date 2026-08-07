// Package service ports org.openelisglobal.panel.service.
// Folder layout mirrors the Java source during migration.
package service

import (
	"openelis-go/internal/panel/daoimpl"
	"openelis-go/internal/panel/valueholder"
)

// PanelService ports PanelServiceImpl — delegates to the DAO.
type PanelService struct {
	DAO *daoimpl.PanelDAOImpl
}

// GetAllActivePanels mirrors PanelServiceImpl.getAllActivePanels().
func (s *PanelService) GetAllActivePanels() ([]valueholder.Panel, error) {
	return s.DAO.GetAllActivePanels()
}
