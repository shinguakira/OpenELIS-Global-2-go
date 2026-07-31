// Package rest ports org.openelisglobal.testconfiguration.controller.rest.TestCatalogRestController
// — GET /rest/TestCatalog (read-only; admin-gated in Java but open in Go during migration).
// Folder layout mirrors the Java source during migration.
package rest

import (
	"database/sql"
	"math"
	"net/http"
	"sort"

	"openelis-go/internal/common/web"
)

// localizationValueEntry mirrors LocalizationValue (with @JsonIgnore on back-ref).
type localizationValueEntry struct {
	ID     string `json:"id"`
	Locale string `json:"locale"`
	Value  string `json:"value"`
}

// localizationEntry mirrors Localization serialized by Jackson (NON_NULL + localization_value map).
type localizationEntry struct {
	ID          string                             `json:"id"`
	Description *string                            `json:"description,omitempty"`
	Values      map[string]localizationValueEntry  `json:"values"`
	Lastupdated *int64                             `json:"lastupdated,omitempty"`
}

// testCatalogItem mirrors the fields of TestCatalog that the e2e parity spec verifies.
// Additional fields (testUnit, sampleType, panel, resultType, …) are included so the
// response is functionally useful to the frontend.
type testCatalogItem struct {
	ID               string            `json:"id"`
	Localization     localizationEntry `json:"localization"`
	TestUnit         string            `json:"testUnit"`
	SampleType       string            `json:"sampleType"`
	Panel            string            `json:"panel"`
	ResultType       string            `json:"resultType"`
	Active           string            `json:"active"`
	Orderable        string            `json:"orderable"`
	Loinc            *string           `json:"loinc,omitempty"`
	Uom              string            `json:"uom"`
	SignificantDigits string           `json:"significantDigits"`
	HasLimitValues   bool              `json:"hasLimitValues"`
	HasDictionaryValues bool           `json:"hasDictionaryValues"`
}

// testCatalogForm mirrors TestCatalogForm (formName set via BaseForm.setFormName).
type testCatalogForm struct {
	FormName        string            `json:"formName"`
	TestCatalogList []testCatalogItem `json:"testCatalogList"`
	TestSectionList []string          `json:"testSectionList"`
}

// testRow is the raw DB row from the test+section query.
type testRow struct {
	testID      string
	sectionName string
	sortOrder   int64
	locID       sql.NullString
	sampleType  string
	active      string
	orderable   string
	loinc       sql.NullString
	uomName     string
}

// TestCatalogService builds the test catalog response from the DB.
type TestCatalogService struct {
	DB *sql.DB
}

// buildForm mirrors TestCatalogRestController.showTestCatalog() + createTestList().
func (s *TestCatalogService) buildForm() (*testCatalogForm, error) {
	// Step 1: fetch all tests with section names and sample types for sorting.
	tests, err := s.fetchTests()
	if err != nil {
		return nil, err
	}
	if len(tests) == 0 {
		return &testCatalogForm{
			FormName:        "testCatalogForm",
			TestCatalogList: []testCatalogItem{},
			TestSectionList: []string{},
		}, nil
	}

	// Step 2: fetch localization values for all test name_localization_ids.
	locIDs := make([]string, 0, len(tests))
	seen := map[string]bool{}
	for _, t := range tests {
		if t.locID.Valid && !seen[t.locID.String] {
			locIDs = append(locIDs, t.locID.String)
			seen[t.locID.String] = true
		}
	}
	locMap, err := s.fetchLocalizations(locIDs)
	if err != nil {
		return nil, err
	}

	// Step 3: sort by (testUnit, sampleType, testSortOrder) — mirrors Java Comparator.
	sort.SliceStable(tests, func(i, j int) bool {
		if tests[i].sectionName != tests[j].sectionName {
			return tests[i].sectionName < tests[j].sectionName
		}
		if tests[i].sampleType != tests[j].sampleType {
			return tests[i].sampleType < tests[j].sampleType
		}
		return tests[i].sortOrder < tests[j].sortOrder
	})

	// Step 4: build testCatalogList and extract unique testSectionList.
	catalogList := make([]testCatalogItem, 0, len(tests))
	sectionList := []string{}
	seenSections := map[string]bool{}

	for _, t := range tests {
		loc := localizationEntry{
			Values: map[string]localizationValueEntry{},
		}
		if t.locID.Valid {
			if entry, ok := locMap[t.locID.String]; ok {
				loc = entry
			}
		}

		active := "Active"
		if t.active != "Y" {
			active = "Not active"
		}
		orderable := "Orderable"
		if t.orderable != "true" {
			orderable = "Not orderable"
		}

		item := testCatalogItem{
			ID:                t.testID,
			Localization:      loc,
			TestUnit:          t.sectionName,
			SampleType:        t.sampleType,
			Panel:             "None",
			ResultType:        "",
			Active:            active,
			Orderable:         orderable,
			Uom:               "n/a",
			SignificantDigits: "n/a",
		}
		if t.loinc.Valid {
			v := t.loinc.String
			item.Loinc = &v
		}
		catalogList = append(catalogList, item)

		if !seenSections[t.sectionName] {
			seenSections[t.sectionName] = true
			sectionList = append(sectionList, t.sectionName)
		}
	}

	return &testCatalogForm{
		FormName:        "testCatalogForm",
		TestCatalogList: catalogList,
		TestSectionList: sectionList,
	}, nil
}

// fetchTests returns all tests with their section names and basic fields for sorting.
func (s *TestCatalogService) fetchTests() ([]testRow, error) {
	rows, err := s.DB.Query(`
		SELECT
		    t.id::text,
		    COALESCE(ts_lv.value, ts.name, '') AS section_name,
		    COALESCE(t.sort_order, $1)::bigint AS sort_order,
		    t.name_localization_id::text,
		    COALESCE(tos_lv.value, tos.description, tos.local_abbrev, 'n/a') AS sample_type,
		    COALESCE(t.is_active, 'N'),
		    COALESCE(t.orderable::text, 'false'),
		    t.loinc,
		    COALESCE(uom.name, 'n/a') AS uom_name
		FROM clinlims.test t
		LEFT JOIN clinlims.test_section ts ON ts.id = t.test_section_id
		LEFT JOIN clinlims.localization_value ts_lv ON (
		    ts_lv.localization_id = ts.name_localization_id AND ts_lv.locale = 'en'
		)
		LEFT JOIN LATERAL (
		    SELECT tost.sample_type_id
		    FROM clinlims.sampletype_test tost
		    WHERE tost.test_id = t.id
		    LIMIT 1
		) first_tost ON true
		LEFT JOIN clinlims.type_of_sample tos ON tos.id = first_tost.sample_type_id
		LEFT JOIN clinlims.localization_value tos_lv ON (
		    tos_lv.localization_id = tos.name_localization_id AND tos_lv.locale = 'en'
		)
		LEFT JOIN clinlims.unit_of_measure uom ON uom.id = t.uom_id`,
		math.MaxInt32)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []testRow
	for rows.Next() {
		var r testRow
		if err := rows.Scan(
			&r.testID,
			&r.sectionName,
			&r.sortOrder,
			&r.locID,
			&r.sampleType,
			&r.active,
			&r.orderable,
			&r.loinc,
			&r.uomName,
		); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

// fetchLocalizations builds a localizationID → localizationEntry map for the given IDs.
func (s *TestCatalogService) fetchLocalizations(ids []string) (map[string]localizationEntry, error) {
	if len(ids) == 0 {
		return map[string]localizationEntry{}, nil
	}

	// Build the IN list via positional parameters.
	placeholders := make([]any, len(ids))
	paramSQL := "$1"
	for i, id := range ids {
		placeholders[i] = id
		if i > 0 {
			paramSQL += ", $" + itoa(i+1)
		}
	}

	q := `SELECT lv.localization_id::text,
		         l.description,
		         EXTRACT(EPOCH FROM l.lastupdated)::bigint * 1000,
		         lv.id::text,
		         lv.locale,
		         lv.value
		  FROM clinlims.localization l
		  JOIN clinlims.localization_value lv ON lv.localization_id = l.id
		  WHERE l.id IN (` + paramSQL + `)`

	rows, err := s.DB.Query(q, placeholders...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[string]localizationEntry{}
	for rows.Next() {
		var locID, lvID, locale, value string
		var description sql.NullString
		var lastupdatedMs sql.NullInt64
		if err := rows.Scan(&locID, &description, &lastupdatedMs, &lvID, &locale, &value); err != nil {
			return nil, err
		}
		entry, ok := result[locID]
		if !ok {
			entry = localizationEntry{
				ID:     locID,
				Values: map[string]localizationValueEntry{},
			}
			if description.Valid {
				d := description.String
				entry.Description = &d
			}
			if lastupdatedMs.Valid {
				ms := lastupdatedMs.Int64
				entry.Lastupdated = &ms
			}
		}
		entry.Values[locale] = localizationValueEntry{
			ID:     lvID,
			Locale: locale,
			Value:  value,
		}
		result[locID] = entry
	}
	return result, rows.Err()
}

// itoa converts an int to its decimal string (avoid importing strconv just for this).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

// Routes registers GET /rest/TestCatalog.
func Routes(mux *http.ServeMux, svc *TestCatalogService) {
	web.Register(mux, "GET", "rest/TestCatalog", func(w http.ResponseWriter, r *http.Request) {
		form, err := svc.buildForm()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, form)
	})
}
