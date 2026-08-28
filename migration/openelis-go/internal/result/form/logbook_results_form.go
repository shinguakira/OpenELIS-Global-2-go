// Package form ports org.openelisglobal.result.form.LogbookResultsForm and the
// lean AccessionResultsRestResponse (constitution.md Layer V).
package form

import (
	"openelis-go/internal/common/util"
	workplanform "openelis-go/internal/workplan/form"
)

// LogbookResultsForm is what rest/LogbookResults returns.
type LogbookResultsForm struct {
	FormName       string `json:"formName"`
	FormMethod     string `json:"formMethod"`
	CancelAction   string `json:"cancelAction"`
	SubmitOnCancel bool   `json:"submitOnCancel"`
	CancelMethod   string `json:"cancelMethod"`

	Paging          workplanform.PagingDTO `json:"paging"`
	AccessionNumber *string                `json:"accessionNumber,omitempty"`
	SinglePatient   bool                   `json:"singlePatient"`
	CurrentDate     string                 `json:"currentDate"`

	DisplayTestMethod bool `json:"displayTestMethod"`
	DisplayTestKit    bool `json:"displayTestKit"`

	TestResult []TestResultRowDTO `json:"testResult"`

	HivKits      []any  `json:"hivKits"`
	SyphilisKits []any  `json:"syphilisKits"`
	Type         string `json:"type"`

	DisplayMethods      bool `json:"displayMethods"`
	DisplayTestSections bool `json:"displayTestSections"`
	SearchByRange       bool `json:"searchByRange"`
	SearchFinished      bool `json:"searchFinished"`
}

// ResultRefDTO is the trimmed `result` object LogbookResults nests in each row.
//
// FIVE keys, not the whole Result entity — the same field on
// rest/accession-results carries the fully-serialised Hibernate graph instead.
// One entity, two serialisations, decided by which associations happen to be
// initialised when Jackson reaches it.
type ResultRefDTO struct {
	IsActive          string `json:"isActive"`
	ID                string `json:"id"`
	SignificantDigits int    `json:"significantDigits"`
	Grouping          int    `json:"grouping"`
	FhirUUIDAsString  string `json:"fhirUuidAsString"`
}

// TestResultRowDTO is one logbook row.
//
// 45 of its fields are the same on every row; only the 23 below them vary. The
// constants are the TestResultItem bean's initialisers, exactly as in the
// workplan rows — this is the same bean rendered by a different builder, which
// is why the two share so much and still differ in what they populate.
type TestResultRowDTO struct {
	AccessionNumber      string `json:"accessionNumber"`
	SequenceNumber       string `json:"sequenceNumber"`
	ShowSampleDetails    bool   `json:"showSampleDetails"`
	IsGroupSeparator     bool   `json:"isGroupSeparator"`
	SampleGroupingNumber int    `json:"sampleGroupingNumber"`

	TestDate     string `json:"testDate"`
	ReceivedDate string `json:"receivedDate"`

	TestMethod      string `json:"testMethod"`
	AnalysisMethod  string `json:"analysisMethod"`
	TestName        string `json:"testName"`
	TestID          string `json:"testId"`
	TestKitInactive bool   `json:"testKitInactive"`

	UpperNormalRange   float64 `json:"upperNormalRange"`
	LowerNormalRange   float64 `json:"lowerNormalRange"`
	UpperAbnormalRange float64 `json:"upperAbnormalRange"`
	LowerAbnormalRange float64 `json:"lowerAbnormalRange"`
	NormalRange        string  `json:"normalRange"`
	LowerCritical      float64 `json:"lowerCritical"`
	HigherCritical     float64 `json:"higherCritical"`
	SignificantDigits  int     `json:"significantDigits"`

	ShadowResultValue string `json:"shadowResultValue"`
	ResultValue       string `json:"resultValue"`
	Technician        string `json:"technician"`
	Reportable        string `json:"reportable"`

	PatientName string `json:"patientName"`
	// A POINTER: accession-results omits patientId entirely while LogbookResults
	// emits it, and an empty string is not the same as an absent key.
	PatientID    *string `json:"patientId,omitempty"`
	SampleItemID string  `json:"sampleItemId"`
	PatientInfo  string  `json:"patientInfo"`
	NationalID   string  `json:"nationalId"`

	UnitsOfMeasure    string `json:"unitsOfMeasure"`
	ResultType        string `json:"resultType"`
	ResultDisplayType string `json:"resultDisplayType"`
	IsModified        bool   `json:"isModified"`

	AnalysisID       string `json:"analysisId"`
	AnalysisStatusID string `json:"analysisStatusId"`
	ResultID         string `json:"resultId"`
	// TWO shapes for one key. LogbookResults nests the five-key reference;
	// accession-results nests the whole serialised entity. Exactly one is set,
	// and they share the JSON name because Java emits the same field either
	// way — which association graph Jackson finds initialised is what decides.
	// ONE key, TWO shapes, so the field is `any`: LogbookResults nests the
	// five-key ResultRefDTO and accession-results nests the whole serialised
	// ResultEntityDTO. Two struct fields sharing a json tag would make
	// encoding/json drop both.
	Result any `json:"result,omitempty"`

	TechnicianSignatureID string `json:"technicianSignatureId"`
	ResultLimitID         string `json:"resultLimitId"`

	DictionaryResults []any  `json:"dictionaryResults"`
	Methods           []any  `json:"methods"`
	Remove            string `json:"remove"`

	Valid                 bool `json:"valid"`
	Normal                bool `json:"normal"`
	NotIncludedInWorkplan bool `json:"notIncludedInWorkplan"`
	ReadOnly              bool `json:"readOnly"`
	ReferredOut           bool `json:"referredOut"`
	ReferralCanceled      bool `json:"referralCanceled"`
	// Emitted only for an analysis that HAS a referral; a plain row omits both
	// rather than carrying empty strings.
	ReferralID        *string `json:"referralId,omitempty"`
	ReferralReasonID  *string `json:"referralReasonId,omitempty"`
	ShadowReferredOut bool    `json:"shadowReferredOut"`
	ShadowRejected    bool    `json:"shadowRejected"`

	MultiSelectResultValues string `json:"multiSelectResultValues"`
	SampleType              string `json:"sampleType"`
	FailedValidation        bool   `json:"failedValidation"`
	Nonconforming           bool   `json:"nonconforming"`
	TestSortOrder           string `json:"testSortOrder"`

	ReflexParentGroup    int            `json:"reflexParentGroup"`
	DisplayResultAsLog   bool           `json:"displayResultAsLog"`
	QualifiedResultValue string         `json:"qualifiedResultValue"`
	HasQualifiedResult   bool           `json:"hasQualifiedResult"`
	Rejected             bool           `json:"rejected"`
	Refer                bool           `json:"refer"`
	ResultFile           map[string]any `json:"resultFile"`
	ResultValueLog       string         `json:"resultValueLog"`

	ServingAsTestGroupIdentifier bool `json:"servingAsTestGroupIdentifier"`
	EqaSample                    bool `json:"eqaSample"`
	ReflexGroup                  bool `json:"reflexGroup"`
	UserChoiceReflex             bool `json:"userChoiceReflex"`
	ChildReflex                  bool `json:"childReflex"`
}

// NewTestResultRow returns a row with the bean initialisers applied.
func NewTestResultRow() TestResultRowDTO {
	return TestResultRowDTO{
		AnalysisMethod:    "MANUAL",
		Reportable:        "N",
		ResultDisplayType: "TEXT",
		DictionaryResults: []any{},
		Methods:           []any{},
		Remove:            "no",
		Valid:             true,
		Normal:            true,
		// An empty OBJECT, not null and not omitted.
		ResultFile:              map[string]any{},
		MultiSelectResultValues: "{}",
		// "--" is the initialiser; a row WITH a result overwrites it with the
		// base-10 logarithm of the value.
		ResultValueLog: "--",
	}
}

// NewLogbookResultsForm returns the envelope literals.
func NewLogbookResultsForm() LogbookResultsForm {
	return LogbookResultsForm{
		FormName:     "LogbookResultsForm",
		FormMethod:   "POST",
		CancelAction: "Home",
		CancelMethod: "POST",
		Paging: workplanform.PagingDTO{
			TotalPages:       "1",
			CurrentPage:      "1",
			SearchTermToPage: []util.IdValuePair{},
		},
		TestResult:          []TestResultRowDTO{},
		HivKits:             []any{},
		SyphilisKits:        []any{},
		DisplayTestMethod:   true,
		DisplayMethods:      true,
		DisplayTestSections: true,
	}
}

// AccessionResultsResponse is rest/accession-results — the ONE lean read in the
// wave, with no Struts form envelope at all.
type AccessionResultsResponse struct {
	AccessionNumber *string `json:"accessionNumber,omitempty"`
	SearchFinished  bool    `json:"searchFinished"`

	// The patient block is flattened onto the ROOT here, where every other
	// endpoint in the wave nests it inside each row. dob is the RAW
	// entered_birth_date text column — the same value AccessionValidation puts
	// in patientInfo and LogbookResults does NOT use, preferring the parsed
	// birth_date.
	//
	// Pointers: they appear only once a search has run.
	FirstName     *string `json:"firstName,omitempty"`
	LastName      *string `json:"lastName,omitempty"`
	DOB           *string `json:"dob,omitempty"`
	Gender        *string `json:"gender,omitempty"`
	ST            *string `json:"st,omitempty"`
	SubjectNumber *string `json:"subjectNumber,omitempty"`
	NationalID    *string `json:"nationalId,omitempty"`

	TestResult []TestResultRowDTO `json:"testResult"`
}
