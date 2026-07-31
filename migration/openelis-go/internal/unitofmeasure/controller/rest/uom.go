// Package rest ports org.openelisglobal.unitofmeasure.controller.rest.UnitOfMeasureRestController
// — GET /rest/uom (read; POST/create remains in Java).
// Folder layout mirrors the Java source during migration.
package rest

import (
	"database/sql"
	"net/http"
	"strings"

	"openelis-go/internal/common/web"
)

// uomRow mirrors the {id, value} map the Java controller returns.
type uomRow struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

// UomService reads clinlims.unit_of_measure and clinlims.uom_type_map.
type UomService struct {
	DB *sql.DB
}

func (s *UomService) getAll() ([]uomRow, error) {
	return s.scan(`SELECT id::text, COALESCE(name, '') FROM clinlims.unit_of_measure`, nil)
}

// getByType mirrors getUnitOfMeasuresByType(uomType):
// JOIN uom_type_map to return only UOMs mapped to the requested type.
func (s *UomService) getByType(uomType string) ([]uomRow, error) {
	return s.scan(`
		SELECT u.id::text, COALESCE(u.name, '')
		FROM clinlims.unit_of_measure u
		JOIN clinlims.uom_type_map m ON m.uom_id = u.id
		WHERE m.uom_type = $1`, []any{uomType})
}

func (s *UomService) scan(q string, args []any) ([]uomRow, error) {
	var rows *sql.Rows
	var err error
	if len(args) == 0 {
		rows, err = s.DB.Query(q)
	} else {
		rows, err = s.DB.Query(q, args...)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []uomRow{}
	for rows.Next() {
		var r uomRow
		if err := rows.Scan(&r.ID, &r.Value); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

// Routes registers GET /rest/uom (with optional ?type= param).
// Mirrors UnitOfMeasureRestController.getUnitOfMeasuresByType().
func Routes(mux *http.ServeMux, svc *UomService) {
	web.Register(mux, "GET", "rest/uom", func(w http.ResponseWriter, r *http.Request) {
		t := strings.TrimSpace(r.URL.Query().Get("type"))
		var list []uomRow
		var err error
		if t != "" {
			list, err = svc.getByType(t)
		} else {
			list, err = svc.getAll()
		}
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, list)
	})
}
