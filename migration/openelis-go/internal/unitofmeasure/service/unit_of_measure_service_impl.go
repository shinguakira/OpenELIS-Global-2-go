// Package service ports org.openelisglobal.unitofmeasure.service.
// Folder layout mirrors the Java source during migration.
//
// Per constitution.md Layer III — DTO shaping belongs here, not the controller.
package service

import (
	"strconv"

	"openelis-go/internal/unitofmeasure/daoimpl"
	"openelis-go/internal/unitofmeasure/form"
	"openelis-go/internal/unitofmeasure/valueholder"
)

// UnitOfMeasureService ports UnitOfMeasureServiceImpl — delegates reads to the
// appropriate DAO depending on whether a type filter is requested, plus the
// DTO shaping that belongs at this layer.
type UnitOfMeasureService struct {
	UomDAO        *daoimpl.UnitOfMeasureDAOImpl
	UomTypeMapDAO *daoimpl.UomTypeMapDAOImpl
}

// GetAll mirrors UnitOfMeasureServiceImpl.getAll() — every UOM, no filter.
func (s *UnitOfMeasureService) GetAll() ([]form.UnitOfMeasureDTO, error) {
	list, err := s.UomDAO.GetAll()
	if err != nil {
		return nil, err
	}
	return toDTOs(list), nil
}

// GetUnitOfMeasuresByType mirrors UnitOfMeasureRestController.getUnitOfMeasuresByType()
// when ?type= is provided — delegates to UomTypeMapDAO.
func (s *UnitOfMeasureService) GetUnitOfMeasuresByType(uomType string) ([]form.UnitOfMeasureDTO, error) {
	list, err := s.UomTypeMapDAO.GetUnitOfMeasuresByType(uomType)
	if err != nil {
		return nil, err
	}
	return toDTOs(list), nil
}

func toDTO(u valueholder.UnitOfMeasure) form.UnitOfMeasureDTO {
	return form.UnitOfMeasureDTO{ID: strconv.FormatInt(u.ID, 10), Value: u.Name}
}

func toDTOs(list []valueholder.UnitOfMeasure) []form.UnitOfMeasureDTO {
	dtos := make([]form.UnitOfMeasureDTO, len(list))
	for i, u := range list {
		dtos[i] = toDTO(u)
	}
	return dtos
}
