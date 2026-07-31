// Package rest ports org.openelisglobal.testcatalog.controller.rest.TestCatalogEditorRestController
// — three read-only reference endpoints: /lab-units, /sample-types, /panels.
// Folder layout mirrors the Java source during migration.
package rest

import (
	"database/sql"
	"net/http"
	"sort"
	"strings"

	"openelis-go/internal/common/web"
)

// idNameOption mirrors LabUnitOption / SampleTypeOption / PanelOption: {id, name}.
type idNameOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// TestCatalogEditorService provides read access to test_section, type_of_sample, panel.
type TestCatalogEditorService struct {
	DB *sql.DB
}

// labUnits mirrors listLabUnits():
// all test sections; English name via localization_value; sorted case-insensitively.
func (s *TestCatalogEditorService) labUnits() ([]idNameOption, error) {
	rows, err := s.DB.Query(`
		SELECT ts.id::text,
		       COALESCE(lv.value, ts.name, '') AS name
		FROM clinlims.test_section ts
		LEFT JOIN clinlims.localization_value lv
		    ON lv.localization_id = ts.name_localization_id
		    AND lv.locale = 'en'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []idNameOption{}
	for rows.Next() {
		var opt idNameOption
		if err := rows.Scan(&opt.ID, &opt.Name); err != nil {
			return nil, err
		}
		list = append(list, opt)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// Java: options.sort((a,b) -> a.name.compareToIgnoreCase(b.name)).
	sort.Slice(list, func(i, j int) bool {
		return strings.ToLower(list[i].Name) < strings.ToLower(list[j].Name)
	})
	return list, nil
}

// sampleTypes mirrors listSampleTypes():
// all type_of_sample ordered by sort_order; name = description || localAbbreviation.
func (s *TestCatalogEditorService) sampleTypes() ([]idNameOption, error) {
	rows, err := s.DB.Query(`
		SELECT id::text,
		       CASE WHEN description IS NOT NULL AND TRIM(description) != ''
		            THEN description
		            ELSE COALESCE(local_abbrev, '') END AS name
		FROM clinlims.type_of_sample
		ORDER BY sort_order`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []idNameOption{}
	for rows.Next() {
		var opt idNameOption
		if err := rows.Scan(&opt.ID, &opt.Name); err != nil {
			return nil, err
		}
		list = append(list, opt)
	}
	return list, rows.Err()
}

// panels mirrors listPanels():
// active panels (is_active='Y') ordered by panel name (SQL ORDER BY NAME).
func (s *TestCatalogEditorService) panels() ([]idNameOption, error) {
	rows, err := s.DB.Query(`
		SELECT id::text, COALESCE(name, '') AS name
		FROM clinlims.panel
		WHERE is_active = 'Y'
		ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []idNameOption{}
	for rows.Next() {
		var opt idNameOption
		if err := rows.Scan(&opt.ID, &opt.Name); err != nil {
			return nil, err
		}
		list = append(list, opt)
	}
	return list, rows.Err()
}

// Routes registers /rest/test-catalog/lab-units, /sample-types, /panels.
func Routes(mux *http.ServeMux, svc *TestCatalogEditorService) {
	web.Register(mux, "GET", "rest/test-catalog/lab-units", func(w http.ResponseWriter, r *http.Request) {
		list, err := svc.labUnits()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, list)
	})
	web.Register(mux, "GET", "rest/test-catalog/sample-types", func(w http.ResponseWriter, r *http.Request) {
		list, err := svc.sampleTypes()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, list)
	})
	web.Register(mux, "GET", "rest/test-catalog/panels", func(w http.ResponseWriter, r *http.Request) {
		list, err := svc.panels()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, list)
	})
}
