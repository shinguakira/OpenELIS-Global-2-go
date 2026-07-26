// Package service ports org.openelisglobal.localization.service.
// Folder layout mirrors the Java source during migration.
package service

import (
	"openelis-go/internal/localization/daoimpl"
	"openelis-go/internal/localization/valueholder"
)

// SupportedLocaleService ports SupportedLocaleServiceImpl — a thin pass-through to
// the DAO (matches the Java service, which just delegates for these reads).
type SupportedLocaleService struct {
	DAO *daoimpl.SupportedLocaleDAO
}

func (s *SupportedLocaleService) GetAll() ([]valueholder.SupportedLocale, error) {
	return s.DAO.GetAll()
}

func (s *SupportedLocaleService) GetAllActive() ([]valueholder.SupportedLocale, error) {
	return s.DAO.GetAllActive()
}

func (s *SupportedLocaleService) GetFallback() (*valueholder.SupportedLocale, error) {
	return s.DAO.GetFallback()
}
