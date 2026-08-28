// Package form ports org.openelisglobal.workplan.form — the WorkplanForm the
// four rest/WorkPlanBy* endpoints all return (constitution.md Layer V).
// Folder layout mirrors the Java source during migration.
package form

import "openelis-go/internal/common/util"

// PagingDTO is the `paging` block every Struts form in this wave carries.
//
// totalPages and currentPage are STRINGS, not numbers — the Java bean holds
// them as String and Jackson writes them as-is.
type PagingDTO struct {
	TotalPages       string             `json:"totalPages"`
	CurrentPage      string             `json:"currentPage"`
	SearchTermToPage []util.IdValuePair `json:"searchTermToPage"`
}

// WorkplanForm is the envelope shared by all four WorkPlanBy* routes.
//
// Every scalar below is "" or false on every observed response, including the
// populated ones: the controllers never set them. They are still emitted —
// this is a bean, not a HashMap, so Include.NON_NULL keeps the empty strings
// and drops nothing. A port that omitted them would break the screens that
// read the envelope.
type WorkplanForm struct {
	FormName       string `json:"formName"`
	FormMethod     string `json:"formMethod"`
	CancelAction   string `json:"cancelAction"`
	SubmitOnCancel bool   `json:"submitOnCancel"`
	CancelMethod   string `json:"cancelMethod"`

	CurrentDate string    `json:"currentDate"`
	Paging      PagingDTO `json:"paging"`

	SelectedSearchID string `json:"selectedSearchID"`
	TestTypeID       string `json:"testTypeID"`
	TestName         string `json:"testName"`
	SearchFinished   bool   `json:"searchFinished"`

	WorkplanTests []TestResultItemDTO `json:"workplanTests"`

	Type         string `json:"type"`
	SearchAction string `json:"searchAction"`
}

// TestResultItemDTO is one workplan row.
//
// 43 of its 49 fields are the same on every row of every one of the four
// endpoints, because the controllers construct a bare TestResultItem and set
// only a handful of properties — everything else is the bean's initialiser.
// They are hardcoded here for exactly that reason, with the Java defaults.
//
// The six that vary are accessionNumber, sampleGroupingNumber, receivedDate,
// testId, patientName and testName. The last two are POINTERS because they are
// not merely empty on some endpoints, they are ABSENT:
//
//	ByTest      sets patientName ("")   , never testName
//	ByPriority  sets patientName ("")   , sets testName
//	ByPanel     never sets patientName  , sets testName
//	BySection   never sets patientName  , sets testName
//
// patientName is "" rather than a name because getPatientName returns "" unless
// configurationName is "Haiti LNSP"; patientInfo likewise unless
// SUBJECT_ON_WORKPLAN is true. Both are deployment config, so this is the
// behaviour here, not a universal constant.
type TestResultItemDTO struct {
	AccessionNumber      string  `json:"accessionNumber"`
	ShowSampleDetails    bool    `json:"showSampleDetails"`
	IsGroupSeparator     bool    `json:"isGroupSeparator"`
	SampleGroupingNumber int     `json:"sampleGroupingNumber"`
	ReceivedDate         string  `json:"receivedDate"`
	TestName             *string `json:"testName,omitempty"`
	TestID               string  `json:"testId"`
	TestKitInactive      bool    `json:"testKitInactive"`

	// util.JavaDouble, not float64. Jackson writes a `double` 40.0 as `40.0`
	// and Go writes `40`; both parse to the same number, so a differ that
	// unmarshals reports parity and only the raw bytes disagree. Six fields ×
	// eighteen rows is a 216-byte Content-Length difference on one response.
	UpperNormalRange   util.JavaDouble `json:"upperNormalRange"`
	LowerNormalRange   util.JavaDouble `json:"lowerNormalRange"`
	UpperAbnormalRange util.JavaDouble `json:"upperAbnormalRange"`
	LowerAbnormalRange util.JavaDouble `json:"lowerAbnormalRange"`
	NormalRange        string          `json:"normalRange"`
	LowerCritical      util.JavaDouble `json:"lowerCritical"`
	HigherCritical     util.JavaDouble `json:"higherCritical"`
	SignificantDigits  int             `json:"significantDigits"`

	Reportable        string  `json:"reportable"`
	PatientName       *string `json:"patientName,omitempty"`
	PatientInfo       string  `json:"patientInfo"`
	UnitsOfMeasure    string  `json:"unitsOfMeasure"`
	ResultDisplayType string  `json:"resultDisplayType"`
	IsModified        bool    `json:"isModified"`

	DictionaryResults []any  `json:"dictionaryResults"`
	Methods           []any  `json:"methods"`
	Remove            string `json:"remove"`

	Valid                 bool   `json:"valid"`
	Normal                bool   `json:"normal"`
	NotIncludedInWorkplan bool   `json:"notIncludedInWorkplan"`
	ReadOnly              bool   `json:"readOnly"`
	ReferredOut           bool   `json:"referredOut"`
	ReferralCanceled      bool   `json:"referralCanceled"`
	ShadowReferredOut     bool   `json:"shadowReferredOut"`
	ShadowRejected        bool   `json:"shadowRejected"`
	ReferralID            string `json:"referralId"`
	ReferralReasonID      string `json:"referralReasonId"`

	FailedValidation bool `json:"failedValidation"`
	Nonconforming    bool `json:"nonconforming"`

	ReflexParentGroup    int    `json:"reflexParentGroup"`
	DisplayResultAsLog   bool   `json:"displayResultAsLog"`
	QualifiedResultValue string `json:"qualifiedResultValue"`
	HasQualifiedResult   bool   `json:"hasQualifiedResult"`
	Rejected             bool   `json:"rejected"`
	Refer                bool   `json:"refer"`
	ResultValueLog       string `json:"resultValueLog"`

	ServingAsTestGroupIdentifier bool `json:"servingAsTestGroupIdentifier"`
	EqaSample                    bool `json:"eqaSample"`
	ReflexGroup                  bool `json:"reflexGroup"`
	UserChoiceReflex             bool `json:"userChoiceReflex"`
	ChildReflex                  bool `json:"childReflex"`
}

// NewTestResultItem returns a row with every constant field already at its Java
// default, so callers set only what their controller sets.
func NewTestResultItem() TestResultItemDTO {
	return TestResultItemDTO{
		ShowSampleDetails: true,
		Reportable:        "N",
		ResultDisplayType: "TEXT",
		// Non-nil so they serialize as [] — the Java bean initialises both.
		DictionaryResults: []any{},
		Methods:           []any{},
		Remove:            "no",
		Valid:             true,
		Normal:            true,
		// "--" is TestResultItem's initialiser for the log value, not a
		// rendering of an absent result.
		ResultValueLog: "--",
	}
}

// NewWorkplanForm returns the envelope with the literals the controllers never
// change.
func NewWorkplanForm() WorkplanForm {
	return WorkplanForm{
		FormName:     "WorkplanForm",
		FormMethod:   "POST",
		CancelAction: "Home",
		CancelMethod: "POST",
		Paging: PagingDTO{
			TotalPages:       "1",
			CurrentPage:      "1",
			SearchTermToPage: []util.IdValuePair{},
		},
		WorkplanTests: []TestResultItemDTO{},
	}
}
