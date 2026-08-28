// Package form ports org.openelisglobal.resultvalidation.form.ResultValidationForm
// (constitution.md Layer V). Folder layout mirrors the Java source.
package form

import (
	"openelis-go/internal/common/util"
	workplanform "openelis-go/internal/workplan/form"
)

// ResultValidationForm is what rest/AccessionValidation returns.
type ResultValidationForm struct {
	FormName       string `json:"formName"`
	FormMethod     string `json:"formMethod"`
	CancelAction   string `json:"cancelAction"`
	SubmitOnCancel bool   `json:"submitOnCancel"`
	CancelMethod   string `json:"cancelMethod"`

	SearchFinished bool                   `json:"searchFinished"`
	Paging         workplanform.PagingDTO `json:"paging"`
	CurrentDate    string                 `json:"currentDate"`

	ResultList []AnalysisItemDTO `json:"resultList"`

	// testSection and testSectionId are DIFFERENT fields. The controller sets
	// testSectionId from the unitType parameter and never touches testSection,
	// which stays "" — so a port that wrote the section into testSection emits
	// the value under the wrong key AND drops the right one.
	TestSection     string  `json:"testSection"`
	TestSectionID   *string `json:"testSectionId,omitempty"`
	AccessionNumber string  `json:"accessionNumber"`
	TestDate        string  `json:"testDate"`
	TestName        string  `json:"testName"`

	// TWO section lists side by side, and they are NOT the same list:
	// testSections is the caller's own sections from getUserTestSections with
	// ROLE_VALIDATION and carries the LOCALIZED name; testSectionsByName is
	// ListType.TEST_SECTION_BY_NAME and carries the RAW test_section.name. They
	// differ wherever a section is stored in one language and localized to
	// another.
	TestSections        []util.IdValuePair `json:"testSections"`
	TestSectionsByName  []util.IdValuePair `json:"testSectionsByName"`
	DisplayTestSections bool               `json:"displayTestSections"`
}

// AnalysisItemDTO is one row awaiting validation.
type AnalysisItemDTO struct {
	Units           string `json:"units"`
	TestName        string `json:"testName"`
	AccessionNumber string `json:"accessionNumber"`
	PatientName     string `json:"patientName"`
	PatientInfo     string `json:"patientInfo"`
	Result          string `json:"result"`

	IsAccepted       bool `json:"isAccepted"`
	IsRejected       bool `json:"isRejected"`
	SampleIsAccepted bool `json:"sampleIsAccepted"`
	SampleIsRejected bool `json:"sampleIsRejected"`

	AnalysisID string `json:"analysisId"`
	TestID     string `json:"testId"`
	ResultID   string `json:"resultId"`

	// Declared `double` in Java, and yet lowerCritical can answer with the JSON
	// STRING "Infinity" — see the switch in the service. util.JavaDouble carries
	// both: a finite value renders 0.0, an infinity renders "Infinity", which is
	// exactly Jackson's own split.
	LowerCritical  util.JavaDouble `json:"lowerCritical"`
	HigherCritical util.JavaDouble `json:"higherCritical"`
	NormalRange    string          `json:"normalRange"`
	ResultType     string          `json:"resultType"`

	IsHighlighted        bool   `json:"isHighlighted"`
	SampleGroupingNumber int    `json:"sampleGroupingNumber"`
	TestSortNumber       string `json:"testSortNumber"`
	DisplayResultAsLog   bool   `json:"displayResultAsLog"`
	ShowAcceptReject     bool   `json:"showAcceptReject"`

	DictionaryResults       []any  `json:"dictionaryResults"`
	MultiSelectResultValues string `json:"multiSelectResultValues"`

	ReadOnly             bool   `json:"readOnly"`
	Nonconforming        bool   `json:"nonconforming"`
	QualifiedResultValue string `json:"qualifiedResultValue"`
	HasQualifiedResult   bool   `json:"hasQualifiedResult"`
	SignificantDigits    int    `json:"significantDigits"`

	Valid                   bool `json:"valid"`
	Positive                bool `json:"positive"`
	Manual                  bool `json:"manual"`
	MultipleResultForSample bool `json:"multipleResultForSample"`
	Normal                  bool `json:"normal"`
	ReflexGroup             bool `json:"reflexGroup"`
	ChildReflex             bool `json:"childReflex"`
}

// NewAnalysisItem returns a row with the bean's initialisers already applied.
func NewAnalysisItem() AnalysisItemDTO {
	return AnalysisItemDTO{
		ShowAcceptReject:        true,
		DictionaryResults:       []any{},
		MultiSelectResultValues: "{}",
		Valid:                   true,
	}
}

// NewResultValidationForm returns the envelope literals.
func NewResultValidationForm() ResultValidationForm {
	return ResultValidationForm{
		FormName:     "ResultValidationForm",
		FormMethod:   "POST",
		CancelAction: "Home",
		CancelMethod: "POST",
		Paging: workplanform.PagingDTO{
			TotalPages:       "1",
			CurrentPage:      "1",
			SearchTermToPage: []util.IdValuePair{},
		},
		ResultList:          []AnalysisItemDTO{},
		DisplayTestSections: true,
	}
}
