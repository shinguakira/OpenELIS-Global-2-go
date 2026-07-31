// Package rest ports org.openelisglobal.dictionary.controller.rest.DictionaryMenuRestController
// — GET /rest/dictionary-categories only (read; writes remain in Java).
// Folder layout mirrors the Java source during migration.
package rest

import (
	"database/sql"
	"net/http"

	"openelis-go/internal/common/web"
)

// dictCategoryRow mirrors DictionaryCategory serialized by Jackson (NON_NULL):
// lastupdated is omitted when the DB row has no timestamp.
type dictCategoryRow struct {
	ID                string `json:"id"`
	Description       string `json:"description"`
	LocalAbbreviation string `json:"localAbbreviation"`
	CategoryName      string `json:"categoryName"`
	Lastupdated       *int64 `json:"lastupdated,omitempty"`
}

// DictionaryCategoryService reads clinlims.dictionary_category.
type DictionaryCategoryService struct {
	DB *sql.DB
}

func (s *DictionaryCategoryService) getAll() ([]dictCategoryRow, error) {
	rows, err := s.DB.Query(`
		SELECT id::text,
		       COALESCE(description, '') AS description,
		       COALESCE(local_abbrev, '') AS local_abbreviation,
		       COALESCE(name, '') AS category_name,
		       EXTRACT(EPOCH FROM lastupdated)::bigint * 1000 AS lastupdated_ms
		FROM clinlims.dictionary_category`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []dictCategoryRow{}
	for rows.Next() {
		var r dictCategoryRow
		var ms sql.NullInt64
		if err := rows.Scan(&r.ID, &r.Description, &r.LocalAbbreviation, &r.CategoryName, &ms); err != nil {
			return nil, err
		}
		if ms.Valid {
			v := ms.Int64
			r.Lastupdated = &v
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

// Routes registers GET /rest/dictionary-categories.
// Mirrors DictionaryMenuRestController.fetchDictionaryCategories().
func Routes(mux *http.ServeMux, svc *DictionaryCategoryService) {
	web.Register(mux, "GET", "rest/dictionary-categories", func(w http.ResponseWriter, r *http.Request) {
		list, err := svc.getAll()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, list)
	})
}
