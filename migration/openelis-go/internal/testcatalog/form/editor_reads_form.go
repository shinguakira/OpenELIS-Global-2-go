package form

// The DTOs of the ten reads that pair with no write.

// TestListRow is one row of the list page — and, reused verbatim, of siblings.
//
// The reuse is where its oddest property comes from. `active`, `amr` and
// `coverageIncomplete` are PRIMITIVE booleans on the Java class, so they
// serialise even when nothing set them. The list page fills all three; siblings
// fills none, so every sibling answers `active: false` regardless of the test's
// real state. Measured, and reproduced.
//
// `coverageIncomplete` is hardcoded false on both paths — the decoration was
// left for a later milestone and never wired.
type TestListRow struct {
	TestID             string  `json:"testId,omitempty"`
	Name               string  `json:"name,omitempty"`
	SampleType         *string `json:"sampleType,omitempty"`
	Code               *string `json:"code,omitempty"`
	Domain             *string `json:"domain,omitempty"`
	Active             bool    `json:"active"`
	AMR                bool    `json:"amr"`
	CoverageIncomplete bool    `json:"coverageIncomplete"`
}

// TestListPage is GET /rest/test-catalog/tests.
//
// `total` is the count AFTER filtering and BEFORE paging, so it is the number
// of rows the filter matched rather than the number returned.
type TestListPage struct {
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
	Total    int           `json:"total"`
	Rows     []TestListRow `json:"rows"`
}

// EditorEnvelope is GET /rest/test-catalog/tests/{testId} — the shell the
// SideNav-routed editor hydrates from.
type EditorEnvelope struct {
	TestID             string   `json:"testId,omitempty"`
	Name               string   `json:"name,omitempty"`
	Code               *string  `json:"code,omitempty"`
	Domain             *string  `json:"domain,omitempty"`
	ApplicableSections []string `json:"applicableSections"`
}

// LocalizationFieldRef ties one editable name field to the localization row
// behind it.
type LocalizationFieldRef struct {
	Field          string `json:"field"`
	LocalizationID string `json:"localizationId"`
}

// LocalizationRefs is GET /rest/test-catalog/tests/{testId}/localization.
type LocalizationRefs struct {
	TestID string                 `json:"testId"`
	Fields []LocalizationFieldRef `json:"fields"`
}

// TestRef is one of the tests a LOINC collides with.
type TestRef struct {
	TestID string `json:"testId,omitempty"`
	Name   string `json:"name,omitempty"`
}

// LoincIntegrity is GET /rest/test-catalog/tests/{testId}/loinc-integrity —
// warnings only, never a block.
//
// `loinc` is absent when the test has none, which is the same state `noLoinc`
// reports; the two disagree in one direction, because `noLoinc` is true only
// for a test that is ALSO active and orderable.
type LoincIntegrity struct {
	Loinc      *string   `json:"loinc,omitempty"`
	Active     bool      `json:"active"`
	NoLoinc    bool      `json:"noLoinc"`
	Duplicates []TestRef `json:"duplicates"`
}

// DictionaryOption is one typeahead hit.
type DictionaryOption struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// GroupTestSummary is one row of GET /rest/test-catalog/group/summary.
type GroupTestSummary struct {
	TestID     string  `json:"testId,omitempty"`
	Name       string  `json:"name,omitempty"`
	Code       *string `json:"code,omitempty"`
	SampleType *string `json:"sampleType,omitempty"`
	Loinc      *string `json:"loinc,omitempty"`
	Active     bool    `json:"active"`
}

// AnalyzerRow is one analyzer that can run a test.
type AnalyzerRow struct {
	AnalyzerID       string  `json:"analyzerId,omitempty"`
	AnalyzerName     *string `json:"analyzerName,omitempty"`
	AnalyzerTestName *string `json:"analyzerTestName,omitempty"`
}

// AnalyzersResponse is GET /rest/test-catalog/tests/{testId}/analyzers.
type AnalyzersResponse struct {
	TestID    string        `json:"testId"`
	Analyzers []AnalyzerRow `json:"analyzers"`
}

// ReflexRow is one reflex rule triggered by this test.
type ReflexRow struct {
	ID               string  `json:"id,omitempty"`
	RuleName         *string `json:"ruleName,omitempty"`
	TriggerCondition string  `json:"triggerCondition,omitempty"`
	ReflexTests      *string `json:"reflexTests,omitempty"`
}

// CalcRow is one calculation touching this test.
type CalcRow struct {
	ID         string  `json:"id,omitempty"`
	Name       *string `json:"name,omitempty"`
	Formula    *string `json:"formula,omitempty"`
	OutputTest *string `json:"outputTest,omitempty"`
}

// ReflexCalcView is GET /rest/test-catalog/{testId}/reflex-calc — read-only
// cross-links. The editor never edits reflex rules or calculations; this only
// surfaces what touches the test.
type ReflexCalcView struct {
	ReflexRules  []ReflexRow `json:"reflexRules"`
	CalculatedBy []CalcRow   `json:"calculatedBy"`
	FeedsInto    []CalcRow   `json:"feedsInto"`
}

// StorageHistoryEntry is one row of GET
// /rest/test-catalog/{testId}/storage/history.
//
// The controller returns the ENTITY, not a DTO, so these are its bean
// properties in Jackson's order — and the two jsonb columns come back as
// STRINGS, because the entity types them as String.
type StorageHistoryEntry struct {
	Lastupdated          *int64  `json:"lastupdated,omitempty"`
	ID                   string  `json:"id,omitempty"`
	TestSampleHandlingID string  `json:"testSampleHandlingId,omitempty"`
	ChangedBy            *string `json:"changedBy,omitempty"`
	ChangedAt            *int64  `json:"changedAt,omitempty"`
	ChangeType           *string `json:"changeType,omitempty"`
	PreviousValues       *string `json:"previousValues,omitempty"`
	NewValues            *string `json:"newValues,omitempty"`
}
