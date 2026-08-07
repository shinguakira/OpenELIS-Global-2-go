// Package service ports org.openelisglobal.typeofsample.service.
// Folder layout mirrors the Java source during migration.
package service

import (
	"openelis-go/internal/typeofsample/daoimpl"
	"openelis-go/internal/typeofsample/valueholder"
)

// TypeOfSampleService ports TypeOfSampleServiceImpl — delegates to the DAO.
type TypeOfSampleService struct {
	DAO *daoimpl.TypeOfSampleDAOImpl
}

// GetAllTypeOfSamplesSortOrdered mirrors TypeOfSampleServiceImpl.getAllTypeOfSamplesSortOrdered().
func (s *TypeOfSampleService) GetAllTypeOfSamplesSortOrdered() ([]valueholder.TypeOfSample, error) {
	return s.DAO.GetAllSortOrdered()
}
