package form

import (
	"openelis-go/internal/common/util"
	locform "openelis-go/internal/localization/form"
)

// ResultLimitBean is one row of a numeric test's limit table, already rendered
// for display — every field is a STRING built by ResultLimitService, not a
// number. A blank gender renders "n/a" and an unbounded age renders "Any Age".
type ResultLimitBean struct {
	Gender         string `json:"gender"`
	AgeRange       string `json:"ageRange"`
	NormalRange    string `json:"normalRange"`
	ValidRange     string `json:"validRange"`
	ReportingRange string `json:"reportingRange"`
	CriticalRange  string `json:"criticalRange"`
}

// TestCatalogBean is one test as TestModifyEntry lists it.
//
// Three fields default to "n/a" rather than to empty — uom, significantDigits
// and a limit's gender — so the screen never has to. The dictionary block and
// the limit block are mutually exclusive and each is absent when its flag is
// false, because the form is NON_NULL.
type TestCatalogBean struct {
	ID                      string                   `json:"id"`
	Localization            *locform.LocalizationDTO `json:"localization,omitempty"`
	ReportLocalization      *locform.LocalizationDTO `json:"reportLocalization,omitempty"`
	TestUnit                string                   `json:"testUnit"`
	SampleType              string                   `json:"sampleType"`
	Panel                   string                   `json:"panel"`
	ResultType              string                   `json:"resultType"`
	Uom                     string                   `json:"uom"`
	SignificantDigits       string                   `json:"significantDigits"`
	Loinc                   *string                  `json:"loinc,omitempty"`
	Active                  string                   `json:"active"`
	Orderable               string                   `json:"orderable"`
	NotifyResults           bool                     `json:"notifyResults"`
	HasDictionaryValues     bool                     `json:"hasDictionaryValues"`
	DictionaryValues        *[]string                `json:"dictionaryValues,omitempty"`
	DictionaryIDs           *[]string                `json:"dictionaryIds,omitempty"`
	ReferenceValue          *string                  `json:"referenceValue,omitempty"`
	ReferenceID             *string                  `json:"referenceId,omitempty"`
	HasLimitValues          bool                     `json:"hasLimitValues"`
	ResultLimits            *[]ResultLimitBean       `json:"resultLimits,omitempty"`
	TestSortOrder           int                      `json:"testSortOrder"`
	InLabOnly               bool                     `json:"inLabOnly"`
	AntimicrobialResistance bool                     `json:"antimicrobialResistance"`
}

// TestModifyEntryForm is GET/POST /rest/TestModifyEntry.
//
// It carries TestAdd's eight lists plus the catalogue, and four name fields the
// GET never fills. labUnitList is the ACTIVE test sections ONLY — TestAdd
// concatenates the inactive ones onto it and this screen does not, so a test
// cannot be moved onto a disabled section from here.
type TestModifyEntryForm struct {
	FormName       string `json:"formName"`
	FormMethod     string `json:"formMethod"`
	CancelAction   string `json:"cancelAction"`
	SubmitOnCancel bool   `json:"submitOnCancel"`
	CancelMethod   string `json:"cancelMethod"`

	NameEnglish       string  `json:"nameEnglish"`
	NameFrench        string  `json:"nameFrench"`
	ReportNameEnglish string  `json:"reportNameEnglish"`
	ReportNameFrench  string  `json:"reportNameFrench"`
	TestID            string  `json:"testId"`
	Loinc             *string `json:"loinc,omitempty"`
	JSONWad           string  `json:"jsonWad"`

	SampleTypeList        *[]util.IdValuePair   `json:"sampleTypeList,omitempty"`
	PanelList             *[]util.IdValuePair   `json:"panelList,omitempty"`
	UomList               *[]util.IdValuePair   `json:"uomList,omitempty"`
	ResultTypeList        *[]util.IdValuePair   `json:"resultTypeList,omitempty"`
	AgeRangeList          *[]util.IdValuePair   `json:"ageRangeList,omitempty"`
	LabUnitList           *[]util.IdValuePair   `json:"labUnitList,omitempty"`
	DictionaryList        *[]util.IdValuePair   `json:"dictionaryList,omitempty"`
	GroupedDictionaryList *[][]util.IdValuePair `json:"groupedDictionaryList,omitempty"`
	TestCatBeanList       *[]TestCatalogBean    `json:"testCatBeanList,omitempty"`
}

// TestModifyEntryPost is the bound body — the same jsonWad string TestAdd takes,
// with a testId inside it.
type TestModifyEntryPost struct {
	JSONWad *string `json:"jsonWad"`
	Loinc   *string `json:"loinc"`
}

// NewTestModifyEntryForm reproduces `new TestModifyEntryForm()`.
func NewTestModifyEntryForm() TestModifyEntryForm {
	return TestModifyEntryForm{
		FormName: "testModifyEntryForm", FormMethod: "POST",
		CancelAction: "Home", SubmitOnCancel: false, CancelMethod: "POST",
	}
}
