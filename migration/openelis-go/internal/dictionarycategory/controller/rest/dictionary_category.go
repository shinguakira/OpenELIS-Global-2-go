// Package rest ports org.openelisglobal.dictionary.controller.rest.DictionaryMenuRestController
// — GET /rest/dictionary-categories only (read; writes remain in Java).
// Folder layout mirrors the Java source during migration.
package rest

import (
	"net/http"

	"openelis-go/internal/common/web"
	"openelis-go/internal/dictionarycategory/service"
	"openelis-go/internal/dictionarycategory/valueholder"
)

// dictCategoryDTO is the JSON shape for each DictionaryCategory row.
// lastupdated is omitted when nil: mirrors Jackson @JsonInclude(NON_NULL).
type dictCategoryDTO struct {
	ID                string `json:"id"`
	Description       string `json:"description"`
	LocalAbbreviation string `json:"localAbbreviation"`
	CategoryName      string `json:"categoryName"`
	Lastupdated       *int64 `json:"lastupdated,omitempty"`
}

func toDTO(c valueholder.DictionaryCategory) dictCategoryDTO {
	return dictCategoryDTO{
		ID:                c.ID,
		Description:       c.Description,
		LocalAbbreviation: c.LocalAbbreviation,
		CategoryName:      c.CategoryName,
		Lastupdated:       c.Lastupdated,
	}
}

// DictionaryMenuRestController mirrors DictionaryMenuRestController.
type DictionaryMenuRestController struct {
	Service *service.DictionaryCategoryService
}

// Routes registers GET /rest/dictionary-categories.
// Mirrors DictionaryMenuRestController.fetchDictionaryCategories().
func Routes(mux *http.ServeMux, ctrl *DictionaryMenuRestController) {
	web.Register(mux, "GET", "rest/dictionary-categories", func(w http.ResponseWriter, r *http.Request) {
		categories, err := ctrl.Service.GetAll()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		dtos := make([]dictCategoryDTO, len(categories))
		for i, c := range categories {
			dtos[i] = toDTO(c)
		}
		web.WriteJSON(w, http.StatusOK, dtos)
	})
}
