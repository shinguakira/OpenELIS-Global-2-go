// Package rest ports org.openelisglobal.testconfiguration.controller.rest.TestCatalogRestController
// — GET /rest/TestCatalog (read-only; admin-gated in Java but open in Go during migration).
// Folder layout mirrors the Java source during migration.
package rest

import (
	"net/http"

	"openelis-go/internal/common/web"
	locvh "openelis-go/internal/localization/valueholder"
	testvh "openelis-go/internal/test/valueholder"
	"openelis-go/internal/testconfiguration/form"
	"openelis-go/internal/testconfiguration/service"
)

// --- DTO types (JSON-tagged; kept in the controller layer) ---

type localizationValueDTO struct {
	ID     string `json:"id"`
	Locale string `json:"locale"`
	Value  string `json:"value"`
}

type localizationDTO struct {
	ID          string                          `json:"id"`
	Description *string                         `json:"description,omitempty"`
	Values      map[string]localizationValueDTO  `json:"values"`
	Lastupdated *int64                          `json:"lastupdated,omitempty"`
}

type testCatalogItemDTO struct {
	ID                  string          `json:"id"`
	Localization        localizationDTO `json:"localization"`
	TestUnit            string          `json:"testUnit"`
	SampleType          string          `json:"sampleType"`
	Panel               string          `json:"panel"`
	ResultType          string          `json:"resultType"`
	Active              string          `json:"active"`
	Orderable           string          `json:"orderable"`
	Loinc               *string         `json:"loinc,omitempty"`
	Uom                 string          `json:"uom"`
	SignificantDigits    string          `json:"significantDigits"`
	HasLimitValues      bool            `json:"hasLimitValues"`
	HasDictionaryValues bool            `json:"hasDictionaryValues"`
}

type testCatalogFormDTO struct {
	FormName        string               `json:"formName"`
	TestCatalogList []testCatalogItemDTO `json:"testCatalogList"`
	TestSectionList []string             `json:"testSectionList"`
}

// --- DTO converters ---

func toLocalizationDTO(loc locvh.Localization) localizationDTO {
	vals := make(map[string]localizationValueDTO, len(loc.Values))
	for locale, lv := range loc.Values {
		vals[locale] = localizationValueDTO{ID: lv.ID, Locale: lv.Locale, Value: lv.Value}
	}
	dto := localizationDTO{
		ID:          loc.ID,
		Values:      vals,
		Lastupdated: loc.Lastupdated,
	}
	if loc.Description != "" {
		d := loc.Description
		dto.Description = &d
	}
	return dto
}

func toCatalogItemDTO(item testvh.TestCatalog) testCatalogItemDTO {
	return testCatalogItemDTO{
		ID:                  item.ID,
		Localization:        toLocalizationDTO(item.Localization),
		TestUnit:            item.TestUnit,
		SampleType:          item.SampleType,
		Panel:               item.Panel,
		ResultType:          item.ResultType,
		Active:              item.Active,
		Orderable:           item.Orderable,
		Loinc:               item.Loinc,
		Uom:                 item.Uom,
		SignificantDigits:    item.SignificantDigits,
		HasLimitValues:      item.HasLimitValues,
		HasDictionaryValues: item.HasDictionaryValues,
	}
}

func toFormDTO(f *form.TestCatalogForm) testCatalogFormDTO {
	items := make([]testCatalogItemDTO, len(f.TestCatalogList))
	for i, item := range f.TestCatalogList {
		items[i] = toCatalogItemDTO(item)
	}
	return testCatalogFormDTO{
		FormName:        f.FormName,
		TestCatalogList: items,
		TestSectionList: f.TestSectionList,
	}
}

// --- Controller ---

// TestCatalogRestController mirrors TestCatalogRestController.
type TestCatalogRestController struct {
	Service *service.TestCatalogService
}

// Routes registers GET /rest/TestCatalog.
func Routes(mux *http.ServeMux, ctrl *TestCatalogRestController) {
	web.Register(mux, "GET", "rest/TestCatalog", func(w http.ResponseWriter, r *http.Request) {
		f, err := ctrl.Service.BuildForm()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, toFormDTO(f))
	})
}
