// Entity-graph DTOs for rest/accession-results.
//
// That endpoint nests the FULLY SERIALISED Hibernate Result entity under each
// row's `result` key — analysis, sample item, sample, type of sample, test
// section, test, unit of measure, panel and every one of their Localization
// objects. rest/LogbookResults nests the SAME field trimmed to five keys.
//
// The difference is not a view annotation: it is which associations happen to
// be initialised when Jackson reaches the object. Reproducing it means
// reproducing the object graph, so these types exist to mirror the Java
// valueholders as Jackson sees them rather than to model anything the port
// needs for itself.
package form

// LocalizedValueDTO is one Localization.values entry.
type LocalizedValueDTO struct {
	Lastupdated *int64 `json:"lastupdated,omitempty"`
	// omitempty: the SYNTHESISED unit-of-measure localization builds its two
	// values without ids, so those two objects carry only locale and value.
	ID     string `json:"id,omitempty"`
	Locale string `json:"locale"`
	Value  string `json:"value"`
}

// LocalizationDTO mirrors the Localization valueholder.
//
// Ten of its thirteen keys are DERIVED getters over the same two rows —
// english/french, valuesAsMap, four locale lists and a rendered
// "English: x" / "French: y" pair. Jackson serialises every public getter, so
// they all reach the wire.
type LocalizationDTO struct {
	Lastupdated *int64                       `json:"lastupdated,omitempty"`
	ID          string                       `json:"id"`
	Description string                       `json:"description"`
	Values      map[string]LocalizedValueDTO `json:"values"`

	LocalizedValue   string   `json:"localizedValue"`
	LocalesWithValue []string `json:"localesWithValue"`
	English          string   `json:"english"`
	French           string   `json:"french"`

	ValuesAsMap                         map[string]string `json:"valuesAsMap"`
	AllActiveLocales                    []string          `json:"allActiveLocales"`
	LocalesWithValueSortedForDisplay    []string          `json:"localesWithValueSortedForDisplay"`
	LocalesSortedForDisplay             []string          `json:"localesSortedForDisplay"`
	LocalesAndValuesOfLocalesWithValues []string          `json:"localesAndValuesOfLocalesWithValues"`
}

// TypeOfSampleDTO mirrors the type_of_sample valueholder.
//
// isActive and active are the SAME boolean under two keys — the field and its
// is-prefixed getter both serialise.
type TypeOfSampleDTO struct {
	Lastupdated       *int64           `json:"lastupdated,omitempty"`
	ID                string           `json:"id"`
	Description       string           `json:"description"`
	Domain            string           `json:"domain"`
	LocalAbbreviation string           `json:"localAbbreviation"`
	IsActive          bool             `json:"isActive"`
	SortOrder         int              `json:"sortOrder"`
	Localization      *LocalizationDTO `json:"localization,omitempty"`
	Active            bool             `json:"active"`
}

// TestSectionDTO mirrors the test_section valueholder. isActive is the CHAR
// column here ("Y"), not a boolean — unlike TypeOfSample's.
type TestSectionDTO struct {
	Lastupdated     *int64           `json:"lastupdated,omitempty"`
	IsActive        string           `json:"isActive"`
	ID              string           `json:"id"`
	IsExternal      string           `json:"isExternal"`
	TestSectionName string           `json:"testSectionName"`
	Description     string           `json:"description"`
	SortOrderInt    int              `json:"sortOrderInt"`
	Localization    *LocalizationDTO `json:"localization,omitempty"`
}

// UnitOfMeasureDTO mirrors the unit_of_measure valueholder.
type UnitOfMeasureDTO struct {
	Lastupdated       *int64           `json:"lastupdated,omitempty"`
	Name              string           `json:"name"`
	Key               string           `json:"key"`
	IsActive          string           `json:"isActive"`
	ID                string           `json:"id"`
	UnitOfMeasureName string           `json:"unitOfMeasureName"`
	Description       string           `json:"description"`
	Localization      *LocalizationDTO `json:"localization,omitempty"`
}

// PanelDTO mirrors the panel valueholder.
type PanelDTO struct {
	Lastupdated  *int64           `json:"lastupdated,omitempty"`
	IsActive     string           `json:"isActive"`
	ID           string           `json:"id"`
	PanelName    string           `json:"panelName"`
	Description  string           `json:"description"`
	SortOrderInt int              `json:"sortOrderInt"`
	Localization *LocalizationDTO `json:"localization,omitempty"`
}

// TestEntityDTO mirrors the test valueholder.
//
// Five of its keys are the same name rendered differently: name, description,
// alternateTestDisplayValue, testDisplayValue and augmentedTestName each
// combine the localized name and the sample type in their own way.
type TestEntityDTO struct {
	Lastupdated               *int64            `json:"lastupdated,omitempty"`
	Name                      string            `json:"name"`
	SortOrder                 string            `json:"sortOrder"`
	Key                       string            `json:"key"`
	IsActive                  string            `json:"isActive"`
	ID                        string            `json:"id"`
	TestSection               *TestSectionDTO   `json:"testSection,omitempty"`
	Description               string            `json:"description"`
	NormalizedDescription     string            `json:"normalizedDescription"`
	Domain                    string            `json:"domain"`
	AlternateTestDisplayValue string            `json:"alternateTestDisplayValue"`
	IsReportable              string            `json:"isReportable"`
	UnitOfMeasure             *UnitOfMeasureDTO `json:"unitOfMeasure,omitempty"`
	LocalCode                 string            `json:"localCode"`
	Orderable                 bool              `json:"orderable"`
	LocalizedTestName         *LocalizationDTO  `json:"localizedTestName,omitempty"`
	LocalizedReportingName    *LocalizationDTO  `json:"localizedReportingName,omitempty"`
	GUID                      string            `json:"guid"`
	InLabOnly                 bool              `json:"inLabOnly"`
	NotifyResults             bool              `json:"notifyResults"`
	AntimicrobialResistance   bool              `json:"antimicrobialResistance"`
	Active                    bool              `json:"active"`
	TestDisplayValue          string            `json:"testDisplayValue"`
	AugmentedTestName         string            `json:"augmentedTestName"`
}

// SampleEntityDTO mirrors the sample valueholder. Note the FOUR renderings of
// the received timestamp: receivedTimestamp and receivedDate as epoch millis,
// receivedDateForDisplay as dd/MM/yyyy, and both receivedTimeForDisplay and
// received24HourTimeForDisplay as HH:mm.
type SampleEntityDTO struct {
	Lastupdated *int64 `json:"lastupdated,omitempty"`
	IsActive    string `json:"isActive"`
	ID          string `json:"id"`

	AccessionNumber       string `json:"accessionNumber"`
	EnteredDate           *int64 `json:"enteredDate,omitempty"`
	EnteredDateForDisplay string `json:"enteredDateForDisplay"`

	ReceivedTimestamp        *int64 `json:"receivedTimestamp,omitempty"`
	ReceivedDateForDisplay   string `json:"receivedDateForDisplay"`
	ReceivedTimeForDisplay   string `json:"receivedTimeForDisplay"`
	CollectionDate           *int64 `json:"collectionDate,omitempty"`
	CollectionDateForDisplay string `json:"collectionDateForDisplay"`
	CollectionTimeForDisplay string `json:"collectionTimeForDisplay"`

	IsConfirmation bool   `json:"isConfirmation"`
	Priority       string `json:"priority"`
	StorageSkipped bool   `json:"storageSkipped"`
	SampleProjects []any  `json:"sampleProjects"`
	StatusID       string `json:"statusId"`
	ObjectID       string `json:"objectId"`

	ReceivedDate                 *int64 `json:"receivedDate,omitempty"`
	Received24HourTimeForDisplay string `json:"received24HourTimeForDisplay"`
	FhirUUIDAsString             string `json:"fhirUuidAsString"`
	BoundTo                      string `json:"boundTo"`
	TableID                      string `json:"tableId"`
}

// SampleItemEntityDTO mirrors the sample_item valueholder.
type SampleItemEntityDTO struct {
	Lastupdated *int64           `json:"lastupdated,omitempty"`
	ID          string           `json:"id"`
	Sample      *SampleEntityDTO `json:"sample,omitempty"`
	SortOrder   string           `json:"sortOrder"`

	TypeOfSample   *TypeOfSampleDTO `json:"typeOfSample,omitempty"`
	TypeOfSampleID string           `json:"typeOfSampleId"`

	CollectionDate *int64 `json:"collectionDate,omitempty"`
	StatusID       string `json:"statusId"`
	Rejected       bool   `json:"rejected"`
	Voided         bool   `json:"voided"`
	// null, not [] — the association is never initialised, and Jackson writes
	// the null rather than dropping it because the getter exists.
	ChildAliquots    *[]any `json:"childAliquots"`
	ObjectID         string `json:"objectId"`
	NestingLevel     int    `json:"nestingLevel"`
	FhirUUIDAsString string `json:"fhirUuidAsString"`
	BoundTo          string `json:"boundTo"`
	TableID          string `json:"tableId"`
	Aliquot          bool   `json:"aliquot"`
}

// AnalysisEntityDTO mirrors the analysis valueholder.
type AnalysisEntityDTO struct {
	Lastupdated  *int64               `json:"lastupdated,omitempty"`
	ID           string               `json:"id"`
	SampleItem   *SampleItemEntityDTO `json:"sampleItem,omitempty"`
	AnalysisType string               `json:"analysisType"`
	TestSection  *TestSectionDTO      `json:"testSection,omitempty"`
	Test         *TestEntityDTO       `json:"test,omitempty"`
	Revision     string               `json:"revision"`
	EnteredDate  *int64               `json:"enteredDate,omitempty"`
	IsReportable string               `json:"isReportable"`
	Panel        *PanelDTO            `json:"panel,omitempty"`

	TriggeredReflex             bool   `json:"triggeredReflex"`
	StatusID                    string `json:"statusId"`
	ReferredOut                 bool   `json:"referredOut"`
	CorrectedSincePatientReport bool   `json:"correctedSincePatientReport"`
	ObjectID                    string `json:"objectId"`
	FhirUUIDAsString            string `json:"fhirUuidAsString"`
	BoundTo                     string `json:"boundTo"`
	TableID                     string `json:"tableId"`
}

// ResultEntityDTO is the FULL Result entity accession-results nests, as opposed
// to the five-key ResultRefDTO LogbookResults nests for the same field.
type ResultEntityDTO struct {
	Lastupdated       *int64             `json:"lastupdated,omitempty"`
	SortOrder         string             `json:"sortOrder"`
	IsActive          string             `json:"isActive"`
	ID                string             `json:"id"`
	Analysis          *AnalysisEntityDTO `json:"analysis,omitempty"`
	IsReportable      string             `json:"isReportable"`
	ResultType        string             `json:"resultType"`
	Value             string             `json:"value"`
	MinNormal         *float64           `json:"minNormal,omitempty"`
	MaxNormal         *float64           `json:"maxNormal,omitempty"`
	SignificantDigits int                `json:"significantDigits"`
	Grouping          int                `json:"grouping"`
	FhirUUIDAsString  string             `json:"fhirUuidAsString"`
}
