// Package service ports org.openelisglobal.localization.service.
// Folder layout mirrors the Java source during migration.
//
// Per constitution.md Layer III — DTO shaping belongs here, not the controller.
package service

import (
	"strconv"

	"openelis-go/internal/localization/daoimpl"
	"openelis-go/internal/localization/form"
	"openelis-go/internal/localization/valueholder"
)

// SupportedLocaleService ports SupportedLocaleServiceImpl — a thin pass-through to
// the DAO (matches the Java service, which just delegates for these reads),
// plus the DTO shaping that belongs at this layer.
type SupportedLocaleService struct {
	DAO *daoimpl.SupportedLocaleDAO
}

func (s *SupportedLocaleService) GetAll() ([]form.SupportedLocaleDTO, error) {
	list, err := s.DAO.GetAll()
	if err != nil {
		return nil, err
	}
	return toDTOs(list), nil
}

func (s *SupportedLocaleService) GetAllActive() ([]form.SupportedLocaleDTO, error) {
	list, err := s.DAO.GetAllActive()
	if err != nil {
		return nil, err
	}
	return toDTOs(list), nil
}

func (s *SupportedLocaleService) GetFallback() (*form.SupportedLocaleDTO, error) {
	fb, err := s.DAO.GetFallback()
	if err != nil {
		return nil, err
	}
	if fb == nil {
		return nil, nil
	}
	dto := toDTO(*fb)
	return &dto, nil
}

func toDTO(l valueholder.SupportedLocale) form.SupportedLocaleDTO {
	return form.SupportedLocaleDTO{
		Id:          strconv.FormatInt(l.Id, 10),
		LocaleCode:  l.LocaleCode,
		DisplayName: l.DisplayName,
		Active:      l.Active,
		Fallback:    l.Fallback,
		SortOrder:   l.SortOrder,
	}
}

func toDTOs(list []valueholder.SupportedLocale) []form.SupportedLocaleDTO {
	dtos := make([]form.SupportedLocaleDTO, 0, len(list)) // non-nil → serializes as []
	for _, l := range list {
		dtos = append(dtos, toDTO(l))
	}
	return dtos
}
