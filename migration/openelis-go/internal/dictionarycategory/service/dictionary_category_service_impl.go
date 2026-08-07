// Package service ports org.openelisglobal.dictionarycategory.service.
// Folder layout mirrors the Java source during migration.
package service

import (
	"openelis-go/internal/dictionarycategory/daoimpl"
	"openelis-go/internal/dictionarycategory/valueholder"
)

// DictionaryCategoryService ports DictionaryCategoryServiceImpl — thin pass-through
// to the DAO (matches Java, which just delegates for these reads).
type DictionaryCategoryService struct {
	DAO *daoimpl.DictionaryCategoryDAOImpl
}

// GetAll mirrors DictionaryCategoryServiceImpl.getAll().
func (s *DictionaryCategoryService) GetAll() ([]valueholder.DictionaryCategory, error) {
	return s.DAO.GetAll()
}
