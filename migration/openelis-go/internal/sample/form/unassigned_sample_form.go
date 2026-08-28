package form

// UnassignedSampleDTO mirrors the HashMap
// UnassignedSampleServiceImpl.compileSampleData builds.
//
// It is a HashMap, so Jackson's Include.NON_NULL applies: a key whose value
// was never put (or was put as null) is ABSENT from the JSON, not null. Hence
// omitempty on every conditionally-populated field. `id`, `priority` and
// `daysUnassigned` are always put, so they never omit.
type UnassignedSampleDTO struct {
	ID             string  `json:"id"`
	ReferralDate   *string `json:"referralDate,omitempty"`
	Priority       string  `json:"priority"`
	DaysUnassigned int64   `json:"daysUnassigned"`

	AccessionNumber *string `json:"accessionNumber,omitempty"`
	SampleID        *string `json:"sampleId,omitempty"`

	ReferralTestName *string `json:"referralTestName,omitempty"`
	TestID           *string `json:"testId,omitempty"`

	DestinationFacilityName *string `json:"destinationFacilityName,omitempty"`
	DestinationFacilityID   *string `json:"destinationFacilityId,omitempty"`

	ReferralReasonID *string `json:"referralReasonId,omitempty"`
}

// CountDTO mirrors the one-key map countUnassignedSamplesByFacility returns:
// HashMap with "count" only.
type CountDTO struct {
	Count int `json:"count"`
}

// UnassignedSampleItemDTO mirrors the subset of SampleItemDTO that the
// unassigned-items queries populate. The Java class has many more fields; they
// stay null on this path and Include.NON_NULL drops them, so declaring them
// here would emit keys Java does not.
//
// The three list fields are initialised to empty lists on the Java object, so
// they DO appear as [] — they are not omitted.
type UnassignedSampleItemDTO struct {
	ID                    string  `json:"id"`
	SampleAccessionNumber string  `json:"sampleAccessionNumber"`
	SampleType            string  `json:"sampleType"`
	SampleTypeID          *string `json:"sampleTypeId,omitempty"`
	CollectionDate        *string `json:"collectionDate,omitempty"`

	ChildAliquots []any `json:"childAliquots"`
	OrderedTests  []any `json:"orderedTests"`
	ReferralTests []any `json:"referralTests"`

	// A primitive boolean on the Java class, so it always serializes.
	HasRemainingQuantity bool `json:"hasRemainingQuantity"`
}
