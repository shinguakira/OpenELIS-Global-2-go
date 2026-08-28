package form

import "openelis-go/internal/sample/daoimpl"

// OrderSearchDTO mirrors the response map OrderSearchRestController.searchOrder
// builds. It is a HashMap, so Include.NON_NULL drops any key put as null —
// hence omitempty on collectionDate, status, patientProperties and orderData,
// each of which Java either puts as null or does not put at all.
type OrderSearchDTO struct {
	ID             string  `json:"id"`
	LabNumber      string  `json:"labNumber"`
	ReceivedDate   string  `json:"receivedDate,omitempty"`
	CollectionDate string  `json:"collectionDate,omitempty"`
	Status         *string `json:"status,omitempty"`

	// Both are put only inside `if (patient != null)`, so a sample with no
	// patient omits them entirely rather than emitting empty objects.
	PatientProperties *PatientInfoBean `json:"patientProperties,omitempty"`
	OrderData         *OrderDataDTO    `json:"orderData,omitempty"`

	Samples          []OrderSearchSampleItemDTO `json:"samples"`
	SampleOrderItems *SampleOrderItemsDTO       `json:"sampleOrderItems"`
	StepProgress     OrderStepProgressDTO       `json:"stepProgress"`
	StorageSkipped   bool                       `json:"storageSkipped"`
}

// OrderDataDTO is the second copy of the patient block Java emits "for frontend
// compatibility", alongside a literal status.
type OrderDataDTO struct {
	PatientProperties   *PatientInfoBean `json:"patientProperties"`
	PatientUpdateStatus string           `json:"patientUpdateStatus"`
}

// PatientInfoBean mirrors org.openelisglobal.patient.action.bean.PatientInfoBean
// as it is populated on this path.
//
// It is a BEAN, not a map, so Include.NON_NULL drops only genuinely null
// fields. Every identity-backed field is set from
// PatientIdentityTypeMap.getIdentityValue, which returns "" (not null) for a
// missing identity — so those keys are always PRESENT and empty, never absent.
// That is why they are plain strings here rather than pointers.
//
// Note `stnumber` is lower-case n in the JSON: the setter is setSTnumber, and
// Jackson's bean naming lower-cases the leading capital run.
type PatientInfoBean struct {
	PatientLastUpdated  string `json:"patientLastUpdated,omitempty"`
	PersonLastUpdated   string `json:"personLastUpdated,omitempty"`
	PatientUpdateStatus string `json:"patientUpdateStatus"`
	PatientPK           string `json:"patientPK"`

	SubjectNumber  string `json:"subjectNumber"`
	NationalID     string `json:"nationalId"`
	GUID           string `json:"guid"`
	LastName       string `json:"lastName"`
	FirstName      string `json:"firstName"`
	AKA            string `json:"aka"`
	MothersName    string `json:"mothersName"`
	MothersInitial string `json:"mothersInitial"`

	City              string `json:"city"`
	Commune           string `json:"commune"`
	AddressDepartment string `json:"addressDepartment"`

	Gender              string `json:"gender"`
	BirthDateForDisplay string `json:"birthDateForDisplay"`

	InsuranceNumber        string `json:"insuranceNumber"`
	Occupation             string `json:"occupation"`
	CustomNotes            string `json:"customNotes"`
	TargetDiseaseProgramme string `json:"targetDiseaseProgramme"`
	PrimaryPhone           string `json:"primaryPhone"`
	Email                  string `json:"email"`
	HealthRegion           string `json:"healthRegion"`
	Education              string `json:"education"`
	MaritialStatus         string `json:"maritialStatus"`
	Nationality            string `json:"nationality"`
	HealthDistrict         string `json:"healthDistrict"`
	OtherNationality       string `json:"otherNationality"`
	STNumber               string `json:"stnumber"`

	// Primitive booleans and an initialised map on the bean, so all three
	// always serialize.
	ReadOnly         bool           `json:"readOnly"`
	IsMerged         bool           `json:"isMerged"`
	AddressHierarchy map[string]any `json:"addressHierarchy"`
}

// SampleOrderItemsDTO mirrors buildSampleOrderItems for the keys this dataset
// exercises. The provider / referring-site / department / program /
// observation-history keys are absent from the live response here and are
// deliberately not built — see the service.
type SampleOrderItemsDTO struct {
	LabNo                  string                `json:"labNo"`
	CollectionDate         string                `json:"collectionDate,omitempty"`
	ReceivedDateForDisplay string                `json:"receivedDateForDisplay,omitempty"`
	ReceivedTime           string                `json:"receivedTime,omitempty"`
	Priority               *string               `json:"priority,omitempty"`
	PaymentOptions         []daoimpl.IDValuePair `json:"paymentOptions"`
}

// OrderSearchSampleItemDTO is one entry of samples[].
type OrderSearchSampleItemDTO struct {
	ID           string  `json:"id"`
	SampleItemID string  `json:"sampleItemId"`
	SortOrder    *string `json:"sortOrder,omitempty"`
	Index        *string `json:"index,omitempty"`
	SampleTypeID *string `json:"sampleTypeId,omitempty"`
	// `name` is typeOfSample.localizedName, `sampleTypeName` its description —
	// both are put only inside `if (typeOfSample != null)`.
	Name           *string `json:"name,omitempty"`
	SampleTypeName *string `json:"sampleTypeName,omitempty"`

	CollectionDate       string   `json:"collectionDate"`
	CollectionTime       string   `json:"collectionTime"`
	ReceivedDate         string   `json:"receivedDate"`
	ReceivedTime         string   `json:"receivedTime"`
	Quantity             *float64 `json:"quantity"`
	QuantityUnit         string   `json:"quantityUnit"`
	CollectorID          string   `json:"collectorId"`
	CollectionConditions string   `json:"collectionConditions"`
	CollectionMethod     string   `json:"collectionMethod"`
	SampleTemperature    string   `json:"sampleTemperature"`
	SpecimenOrigin       string   `json:"specimenOrigin"`

	SampleXML OrderSearchSampleXMLDTO `json:"sampleXML"`
	Tests     []OrderSearchTestDTO    `json:"tests"`
	Panels    []OrderSearchPanelDTO   `json:"panels"`

	// The five storage keys travel together — Java puts them only when there
	// is an assignment WITH a location. storageHierarchicalPath is put only
	// when the ancestry also resolved.
	StorageLocationID         *int64  `json:"storageLocationId,omitempty"`
	StorageLocationType       *string `json:"storageLocationType,omitempty"`
	StoragePositionCoordinate *string `json:"storagePositionCoordinate,omitempty"`
	StorageNotes              *string `json:"storageNotes,omitempty"`
	StorageHierarchicalPath   *string `json:"storageHierarchicalPath,omitempty"`
}

// OrderSearchSampleXMLDTO is the nested duplicate block Java also emits.
// `collector` is put RAW (may be null) here, while the outer `collectorId`
// is coalesced to "" — the same value, two null policies.
type OrderSearchSampleXMLDTO struct {
	CollectionDate    string   `json:"collectionDate"`
	CollectionTime    string   `json:"collectionTime"`
	Quantity          *float64 `json:"quantity"`
	UOM               string   `json:"uom"`
	Collector         *string  `json:"collector,omitempty"`
	CollectionMethod  string   `json:"collectionMethod"`
	SampleTemperature string   `json:"sampleTemperature"`
	SpecimenOrigin    string   `json:"specimenOrigin"`
}

// OrderSearchTestDTO is one entry of a sample item's tests[].
type OrderSearchTestDTO struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

// OrderSearchPanelDTO is one entry of a sample item's panels[].
type OrderSearchPanelDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
