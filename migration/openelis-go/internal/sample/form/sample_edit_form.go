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
	Warning                 bool `json:"warning"`
	CustomNotificationLogic bool `json:"customNotificationLogic"`

	// Accession-number generator configuration. NUMBERS, not strings.
	AccessionFormat      string `json:"accessionFormat"`
	IDSeparator          string `json:"idSeparator"`
	MaxAccessionLength   int    `json:"maxAccessionLength"`
	EditableAccession    int    `json:"editableAccession"`
	NonEditableAccession int    `json:"nonEditableAccession"`

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
	// POINTERS to slices, and both halves of that matter.
	//
	// The controller calls setExistingTests / setAddableTestInfo only inside
	// `if (sample != null)`, so a resolved accession emits `[]` for a sample
	// with no analyses while an unresolved one leaves the field null and
	// NON_NULL drops the key. Neither plain form of a Go slice says that:
	// `omitempty` hides an EMPTY slice as well as a nil one (a resolved sample
	// with no tests lost both keys), and dropping `omitempty` marshals a nil
	// slice as `null` rather than omitting it. A nil POINTER omits; a pointer
	// to an empty slice emits `[]`.
	ExistingTests    *[]SampleEditItemDTO `json:"existingTests,omitempty"`
	PossibleTests    *[]SampleEditItemDTO `json:"possibleTests,omitempty"`
	SampleTypes      []util.IdValuePair   `json:"sampleTypes,omitempty"`
	SampleOrderItems *SampleEditOrderDTO  `json:"sampleOrderItems,omitempty"`
}

// SampleEditItemDTO is one row of existingTests or possibleTests.
//
// The two lists share a class but are populated differently:
//
//   - existingTests rows carry analysisId, status, collectionDate and
//     collectionTime; possibleTests rows do not (they describe a test that
//     could be added, so there is no analysis yet).
//   - The HEADER fields — accessionNumber, sampleType, collectionDate,
//     collectionTime — are set ONCE PER SAMPLE ITEM on BOTH lists: on the first
//     row of each item's group after the sort, left null on the rest, where
//     NON_NULL drops them. canRemoveSample follows the same "first row only"
//     logic but is a primitive boolean, so it stays present as false instead of
//     disappearing.
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
//   - SampleEdit         the form lists PLUS labNo/sampleId/priority AND, when
//     the accession resolves to a sample, the whole
//     observation/requester/program block below
//
// Sharing one builder across the three gets at least one of them wrong.
//
// An earlier revision of this comment said SampleEdit had NO requestDate. It
// does — getBaseSampleOrderItem stamps it from the clock and the sample branch
// overwrites it from observation history. The claim survived because no sample
// in the dataset carried an observation to overwrite it with.
type SampleEditOrderDTO struct {
	LabNo    string `json:"labNo"`
	SampleID string `json:"sampleId"`

	// A POINTER, because getBaseSampleOrderItem never sets priority and the
	// sample branch does `setPriority(sample.getPriority())` — so a NULL
	// order_priority leaves the bean field null and NON_NULL drops the key.
	// Defaulting to "ROUTINE" here (as this port did) invents a value Java
	// does not emit. The column DEFAULTs to 'ROUTINE', so only an explicitly
	// NULLed row shows the difference.
	Priority *string `json:"priority,omitempty"`

	// Stamped from the clock by getBaseSampleOrderItem, then OVERWRITTEN from
	// sample.received_date when the accession resolves. Always emitted.
	ReceivedDateForDisplay string `json:"receivedDateForDisplay"`
	ReceivedTime           string `json:"receivedTime"`

	// Everything below is set only inside `if (sample != null)`, on a bean with
	// Include.NON_NULL — so an unset field drops its key rather than emitting
	// "". Pointers, not strings, for exactly that reason.
	//
	// requestDate is the exception: the base item stamps it from the clock, so
	// it is present even for a sample with no requestDate observation. It is
	// still a pointer because the sample branch may overwrite it with null.
	RequestDate                  *string `json:"requestDate,omitempty"`
	ReferringPatientNumber       *string `json:"referringPatientNumber,omitempty"`
	NextVisitDate                *string `json:"nextVisitDate,omitempty"`
	PaymentOptionSelection       *string `json:"paymentOptionSelection,omitempty"`
	TestLocationCode             *string `json:"testLocationCode,omitempty"`
	OtherLocationCode            *string `json:"otherLocationCode,omitempty"`
	BillingReferenceNumber       *string `json:"billingReferenceNumber,omitempty"`
	ProvisionalClinicalDiagnosis *string `json:"provisionalClinicalDiagnosis,omitempty"`
	Program                      *string `json:"program,omitempty"`
	ProgramID                    *string `json:"programId,omitempty"`

	ProviderPersonID  *string `json:"providerPersonId,omitempty"`
	ProviderID        *string `json:"providerId,omitempty"`
	ProviderFirstName *string `json:"providerFirstName,omitempty"`
	ProviderLastName  *string `json:"providerLastName,omitempty"`
	ProviderWorkPhone *string `json:"providerWorkPhone,omitempty"`
	ProviderFax       *string `json:"providerFax,omitempty"`
	ProviderEmail     *string `json:"providerEmail,omitempty"`

	ReferringSiteID           *string `json:"referringSiteId,omitempty"`
	ReferringSiteName         *string `json:"referringSiteName,omitempty"`
	ReferringSiteCode         *string `json:"referringSiteCode,omitempty"`
	ReferringSiteDepartmentID *string `json:"referringSiteDepartmentId,omitempty"`

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
