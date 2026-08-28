package form

import "openelis-go/internal/common/util"

// SampleEditDTO mirrors org.openelisglobal.sample.form.SampleEditForm as
// SampleEditRestController.showSampleEdit populates it.
//
// It is a BEAN, so Include.NON_NULL drops only genuinely null fields. That
// splits the response three ways and the split is the contract:
//
//   - Scalars initialised on the form ("" or false) are ALWAYS PRESENT, even
//     when nothing was found. patientName, dob, gender, nationalId, patientId,
//     subjectNumber and maxAccessionNumber all come back as "" for an unknown
//     accession rather than disappearing.
//   - Fields only assigned inside the found branch (existingTests,
//     possibleTests, sampleOrderItems, sampleTypes) stay null and VANISH.
//   - accessionNumber is assigned only when one was supplied, so the blank
//     form omits it while an unknown accession echoes it back.
//
// Measured on all three states, not inferred: blank = 29 keys, unknown = 30,
// found = 34.
//
// NOTE the REST controller is not the MVC SampleEditController. It leaves
// `rejectReasonList` and `patientSearch` COMMENTED OUT, so neither key is ever
// emitted here even though the MVC twin sets both. Porting from the MVC class
// would add two keys Java does not send.
type SampleEditDTO struct {
	// Form metadata — hardcoded in the controller.
	FormName   string `json:"formName"`
	FormAction string `json:"formAction"`
	FormMethod string `json:"formMethod"`

	CancelAction   string `json:"cancelAction"`
	CancelMethod   string `json:"cancelMethod"`
	SubmitOnCancel bool   `json:"submitOnCancel"`

	CurrentDate string `json:"currentDate"`
	// Always true on this form; SamplePatientEntry sets it false.
	Warning                  bool `json:"warning"`
	CustomNotificationLogic  bool `json:"customNotificationLogic"`

	// Accession-number generator configuration. NUMBERS, not strings.
	AccessionFormat     string `json:"accessionFormat"`
	IDSeparator         string `json:"idSeparator"`
	MaxAccessionLength  int    `json:"maxAccessionLength"`
	EditableAccession   int    `json:"editableAccession"`
	NonEditableAccession int   `json:"nonEditableAccession"`

	// Search state. Both are primitive booleans, so both always serialize —
	// including the blank form, where BOTH are false.
	SearchFinished bool `json:"searchFinished"`
	NoSampleFound  bool `json:"noSampleFound"`

	// Assigned only when an accessionNumber was supplied.
	AccessionNumber *string `json:"accessionNumber,omitempty"`

	NewAccessionNumber string `json:"newAccessionNumber"`
	MaxAccessionNumber string `json:"maxAccessionNumber"`

	// Patient block: initialised to "" on the form, so present even when empty.
	PatientID     string `json:"patientId"`
	PatientName   string `json:"patientName"`
	DOB           string `json:"dob"`
	Gender        string `json:"gender"`
	NationalID    string `json:"nationalId"`
	SubjectNumber string `json:"subjectNumber"`

	AbleToCancelResults  bool `json:"ableToCancelResults"`
	IsConfirmationSample bool `json:"isConfirmationSample"`
	IsEditable           bool `json:"isEditable"`

	SampleXML string `json:"sampleXML"`

	InitialSampleConditionList []util.IdValuePair `json:"initialSampleConditionList"`

	// Found-branch only — null and therefore absent on the other two states.
	ExistingTests    []SampleEditItemDTO   `json:"existingTests,omitempty"`
	PossibleTests    []SampleEditItemDTO   `json:"possibleTests,omitempty"`
	SampleTypes      []util.IdValuePair    `json:"sampleTypes,omitempty"`
	SampleOrderItems *SampleEditOrderDTO   `json:"sampleOrderItems,omitempty"`
}

// SampleEditItemDTO is one row of existingTests or possibleTests.
//
// The two lists share a class but are populated differently:
//
//   - existingTests rows carry analysisId, status, collectionDate and
//     collectionTime; possibleTests rows do not (they describe a test that
//     could be added, so there is no analysis yet).
//   - accessionNumber and sampleType are set ONCE PER SAMPLE ITEM in
//     possibleTests — on the first row of each item's group, as headers for the
//     frontend — and left null on the rest, where NON_NULL drops them. In
//     existingTests every row has them.
//
// `id` duplicates `testId` on both lists, and `sortOrder` is the TEST's sort
// order rather than the sample item's.
type SampleEditItemDTO struct {
	AccessionNumber *string `json:"accessionNumber,omitempty"`
	SampleType      *string `json:"sampleType,omitempty"`

	AnalysisID     *string `json:"analysisId,omitempty"`
	Status         *string `json:"status,omitempty"`
	CollectionDate *string `json:"collectionDate,omitempty"`
	CollectionTime *string `json:"collectionTime,omitempty"`

	TestName     string `json:"testName"`
	SampleItemID string `json:"sampleItemId"`
	TestID       string `json:"testId"`
	SortOrder    string `json:"sortOrder"`
	ID           string `json:"id"`

	CanCancel         bool `json:"canCancel"`
	Canceled          bool `json:"canceled"`
	Add               bool `json:"add"`
	CanRemoveSample   bool `json:"canRemoveSample"`
	RemoveSample      bool `json:"removeSample"`
	SampleItemChanged bool `json:"sampleItemChanged"`
	HasResults        bool `json:"hasResults"`
}

// SampleEditOrderDTO is the sampleOrderItems variant this form emits.
//
// THIRD distinct object under that key in this migration:
//   - order/search       labNo, collectionDate, priority, paymentOptions
//   - SamplePatientEntry the form LISTS plus requestDate (no labNo — no sample yet)
//   - SampleEdit         the form lists PLUS labNo/sampleId/priority, and NO
//                        requestDate
//
// Sharing one builder across the three gets at least one of them wrong.
type SampleEditOrderDTO struct {
	LabNo    string `json:"labNo"`
	SampleID string `json:"sampleId"`
	Priority string `json:"priority"`

	ReceivedDateForDisplay string `json:"receivedDateForDisplay"`
	ReceivedTime           string `json:"receivedTime"`

	PaymentOptions       []util.IdValuePair `json:"paymentOptions"`
	PriorityList         []util.IdValuePair `json:"priorityList"`
	ProgramList          []util.IdValuePair `json:"programList"`
	ProvidersList        []util.IdValuePair `json:"providersList"`
	ReferringSiteList    []util.IdValuePair `json:"referringSiteList"`
	TestLocationCodeList []util.IdValuePair `json:"testLocationCodeList"`

	EnvironmentalFields map[string]any `json:"environmentalFields"`
	IsEQASample         bool           `json:"isEQASample"`
	Modified            bool           `json:"modified"`
	ReadOnly            bool           `json:"readOnly"`
}
