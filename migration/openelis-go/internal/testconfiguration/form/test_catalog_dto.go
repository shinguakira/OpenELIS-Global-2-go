// Package form ports org.openelisglobal.testconfiguration.form (constitution.md
// Layer V — "Forms/DTOs"). This file holds the JSON-tagged wire shapes; the
// non-JSON-tagged TestCatalogForm in test_catalog_form.go is an internal
// assembly structure used only within the service (Layer III) before it's
// converted to TestCatalogFormDTO — kept separate rather than JSON-tagging
// TestCatalogForm directly, since it's a straight port of Java's own
// TestCatalogForm.java (a form-binding bean, not itself the JSON contract).
package form

type LocalizationValueDTO struct {
	ID     string `json:"id"`
	Locale string `json:"locale"`
	Value  string `json:"value"`
}

type LocalizationDTO struct {
	ID          string                          `json:"id"`
	Description *string                         `json:"description,omitempty"`
	Values      map[string]LocalizationValueDTO `json:"values"`
	Lastupdated *int64                          `json:"lastupdated,omitempty"`
}

type TestCatalogItemDTO struct {
	ID                  string          `json:"id"`
	Localization        LocalizationDTO `json:"localization"`
	TestUnit            string          `json:"testUnit"`
	SampleType          string          `json:"sampleType"`
	Panel               string          `json:"panel"`
	ResultType          string          `json:"resultType"`
	Active              string          `json:"active"`
	Orderable           string          `json:"orderable"`
	Loinc               *string         `json:"loinc,omitempty"`
	Uom                 string          `json:"uom"`
	SignificantDigits   string          `json:"significantDigits"`
	HasLimitValues      bool            `json:"hasLimitValues"`
	HasDictionaryValues bool            `json:"hasDictionaryValues"`
}

type TestCatalogFormDTO struct {
	FormName        string               `json:"formName"`
	TestCatalogList []TestCatalogItemDTO `json:"testCatalogList"`
	TestSectionList []string             `json:"testSectionList"`
}
