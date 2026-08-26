// Package service ports org.openelisglobal.dictionarycategory.service.
// Folder layout mirrors the Java source during migration.
//
// Per constitution.md Layer III (Services) — the "Data Compilation Rule":
// services build the complete response DTO here, not the controller. The
// controller (Layer IV) only does request/response mapping.
package service

import (
	"strconv"

	"openelis-go/internal/dictionarycategory/daoimpl"
	"openelis-go/internal/dictionarycategory/form"
	"openelis-go/internal/dictionarycategory/valueholder"
)

// DictionaryCategoryService ports DictionaryCategoryServiceImpl — thin pass-through
// to the DAO (matches Java, which just delegates for these reads), plus the
// DTO shaping that belongs at this layer.
type DictionaryCategoryService struct {
	DAO *daoimpl.DictionaryCategoryDAOImpl
}

// GetAll mirrors DictionaryCategoryServiceImpl.getAll(), returning the
// ready-to-serialize DTO rows.
func (s *DictionaryCategoryService) GetAll() ([]form.DictionaryCategoryDTO, error) {
	categories, err := s.DAO.GetAll()
	if err != nil {
		return nil, err
	}
	dtos := make([]form.DictionaryCategoryDTO, len(categories))
	for i, c := range categories {
		dtos[i] = toDTO(c)
	}
	return dtos, nil
}

// toDTO converts a DictionaryCategory valueholder to its JSON DTO.
// Lastupdated *time.Time → *int64 (epoch ms) so the JSON output is identical
// to what Jackson produces for Java Date fields.
func toDTO(c valueholder.DictionaryCategory) form.DictionaryCategoryDTO {
	dto := form.DictionaryCategoryDTO{
		ID:                strconv.FormatInt(c.ID, 10),
		Description:       c.Description,
		LocalAbbreviation: c.LocalAbbreviation,
		CategoryName:      c.CategoryName,
	}
	if c.Lastupdated != nil {
		ms := c.Lastupdated.UnixMilli()
		dto.Lastupdated = &ms
	}
	return dto
}
