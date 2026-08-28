package form

// UnassignedSampleDTO mirrors the HashMap
// UnassignedSampleServiceImpl.compileSampleData builds.
//
// It is a HashMap, so Jackson's Include.NON_NULL applies: a key whose value
// was never put (or was put as null) is ABSENT from the JSON, not null. Hence
// omitempty on every conditionally-populated field. `id`, `priority` and
// `daysUnassigned` are always put, so they never omit.
type UnassignedSampleDTO struct {
	ID string `json:"id"`
	// EPOCH MILLIS, not a formatted string. compileSampleData puts the raw
	// java.sql.Timestamp into the map and Jackson serializes it with
	// WRITE_DATES_AS_TIMESTAMPS, so the wire value is a NUMBER
	// (1746086400000), unlike sampleXML.collectionDate and the attachment
	// list's uploadedAt, which are Timestamp.toString() strings. Measured
	// against live Java; this field emitted "2025-05-09 17:00:00.0" until the
	// referral fixture existed to compare against.
	ReferralDate   *int64 `json:"referralDate,omitempty"`
	Priority       string `json:"priority"`
	DaysUnassigned int64  `json:"daysUnassigned"`

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

// There is deliberately NO DTO for rest/unassigned-sample/items here.
//
// Java never serializes one: buildSampleItemDTOs calls
// getReferralsBySampleItemId(Integer) while SampleItem.id is mapped as a
// String, so Hibernate rejects the binding for the FIRST row it reaches, the
// read-only transaction is marked rollback-only, and the commit at the
// @Transactional boundary throws — the controller's catch turns that into a
// bodiless 500. See UnassignedSampleService.GetUnassignedItems.
//
// An earlier version of this file DID declare one, with fields
// (`sampleAccessionNumber`, `sampleType`, `childAliquots`,
// `hasRemainingQuantity`, …) that Java's SampleItemDTO does not even have. It
// was written from a misread of the Java class and nothing could catch it,
// because clinlims.referral was empty and the endpoint answered 200 [] on both
// stacks. Declaring a shape no live response has ever confirmed is exactly the
// guess this migration's e2e discipline exists to prevent, so it is gone
// rather than corrected: if the Java defect is ever fixed, the shape gets
// written from the response that fix produces.
