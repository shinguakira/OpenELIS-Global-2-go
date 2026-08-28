package form

import (
	commonform "openelis-go/internal/common/form"
	"openelis-go/internal/common/util"
	batchform "openelis-go/internal/samplebatchentry/form"
)

// SamplePatientEntryDTO mirrors the SamplePatientEntry form load (Wave 4.6).
//
// Shares a lot of keys with SampleBatchEntrySetup and is NOT the same object.
// The differences that matter:
//   - sampleTypes here is ROLE-FILTERED (getUserSampleTypes with
//     ROLE_RECEPTION); the batch form lists every active human type.
//   - warning is FALSE here and TRUE on SampleEdit — same key, opposite literal
//     from two different controllers.
//   - patientProperties, patientSearch, referralReasons, referralOrganizations
//     and rejectReasonList exist only here.
//   - there is no currentTime, project or projectData block.
type SamplePatientEntryDTO struct {
	FormName   string `json:"formName"`
	FormMethod string `json:"formMethod"`

	CancelAction   string `json:"cancelAction"`
	CancelMethod   string `json:"cancelMethod"`
	SubmitOnCancel bool   `json:"submitOnCancel"`

	CurrentDate string `json:"currentDate"`

	CustomNotificationLogic bool `json:"customNotificationLogic"`
	OrderEntryOnly          bool `json:"orderEntryOnly"`
	UseReferral             bool `json:"useReferral"`
	// FALSE on this form.
	Warning bool `json:"warning"`

	PatientUpdateStatus string `json:"patientUpdateStatus"`
	SampleXML           string `json:"sampleXML"`

	InitialSampleConditionList []util.IdValuePair `json:"initialSampleConditionList"`
	RejectReasonList           []util.IdValuePair `json:"rejectReasonList"`
	ReferralReasons            []util.IdValuePair `json:"referralReasons"`
	ReferralOrganizations      []util.IdValuePair `json:"referralOrganizations"`
	SampleTypes                []util.IdValuePair `json:"sampleTypes"`
	TestSectionList            []util.IdValuePair `json:"testSectionList"`

	Projects []commonform.ProjectEntityDTO `json:"projects"`

	PatientProperties *PatientEntryPropertiesDTO `json:"patientProperties"`
	PatientSearch     *PatientSearchDTO          `json:"patientSearch"`

	// The blank-form variant, shared with SampleBatchEntrySetup.
	SampleOrderItems *batchform.SampleOrderFormDTO `json:"sampleOrderItems"`
}

// PatientEntryPropertiesDTO is the BLANK-FORM patientProperties.
//
// Completely different from order/search's patientProperties, which is a
// populated PatientInfoBean. This one carries the LISTS a new-patient form
// needs and almost no scalars — the two share the key name and nothing else.
//
// healthDistricts and healthRegions are organization-backed and empty on this
// dataset; they are still non-nil so they serialize as [] rather than null.
type PatientEntryPropertiesDTO struct {
	AddressDepartments []commonform.DictionaryEntityDTO `json:"addressDepartments"`
	AddressHierarchy   map[string]any                   `json:"addressHierarchy"`

	// "" because no patient is loaded — present, not dropped, because the bean
	// initialises them.
	BirthDateForDisplay string `json:"birthDateForDisplay"`
	PatientType         string `json:"patientType"`

	EducationList   []util.IdValuePair `json:"educationList"`
	Genders         []util.IdValuePair `json:"genders"`
	MaritialList    []util.IdValuePair `json:"maritialList"`
	NationalityList []util.IdValuePair `json:"nationalityList"`
	HealthDistricts []util.IdValuePair `json:"healthDistricts"`
	HealthRegions   []util.IdValuePair `json:"healthRegions"`

	IDDocuments []any `json:"idDocuments"`

	// Full entities, unlike the {id,value} lists beside them.
	PatientTypes []commonform.PatientTypeEntityDTO `json:"patientTypes"`

	ReadOnly bool `json:"readOnly"`
}

// PatientSearchDTO is the search sub-form.
//
// loadFromServerWithPatient is FALSE here. The MVC SampleEditController sets it
// true on its own PatientSearch, but the REST SampleEdit leaves the whole
// object out — so this is the only endpoint in the wave that emits one.
type PatientSearchDTO struct {
	DefaultHeader             bool               `json:"defaultHeader"`
	Genders                   []util.IdValuePair `json:"genders"`
	LoadFromServerWithPatient bool               `json:"loadFromServerWithPatient"`
	SearchCriteria            []util.IdValuePair `json:"searchCriteria"`
}
