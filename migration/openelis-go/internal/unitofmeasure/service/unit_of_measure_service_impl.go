// Package service ports org.openelisglobal.unitofmeasure.service.
// Folder layout mirrors the Java source during migration.
package service

import (
	"openelis-go/internal/unitofmeasure/daoimpl"
	"openelis-go/internal/unitofmeasure/valueholder"
)

// UnitOfMeasureService ports UnitOfMeasureServiceImpl — delegates reads to the
// appropriate DAO depending on whether a type filter is requested.
type UnitOfMeasureService struct {
	UomDAO        *daoimpl.UnitOfMeasureDAOImpl
	UomTypeMapDAO *daoimpl.UomTypeMapDAOImpl
}

// GetAll mirrors UnitOfMeasureServiceImpl.getAll() — every UOM, no filter.
func (s *UnitOfMeasureService) GetAll() ([]valueholder.UnitOfMeasure, error) {
	return s.UomDAO.GetAll()
}

// GetUnitOfMeasuresByType mirrors UnitOfMeasureRestController.getUnitOfMeasuresByType()
// when ?type= is provided — delegates to UomTypeMapDAO.
func (s *UnitOfMeasureService) GetUnitOfMeasuresByType(uomType string) ([]valueholder.UnitOfMeasure, error) {
	return s.UomTypeMapDAO.GetUnitOfMeasuresByType(uomType)
}
