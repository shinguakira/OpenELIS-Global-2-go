package form

import "openelis-go/internal/common/util"

// The DTOs the editor's test-level endpoints exchange.

// CreateTestRequest is POST /rest/test-catalog/tests — the create-in-place flow.
type CreateTestRequest struct {
	Name          string  `json:"name"`
	ReportingName string  `json:"reportingName"`
	Code          string  `json:"code"`
	LabUnitID     string  `json:"labUnitId"`
	SampleTypeID  string  `json:"sampleTypeId"`
	Domain        string  `json:"domain"`
	AMR           *bool   `json:"amr"`
	Orderable     *bool   `json:"orderable"`
	Description   *string `json:"description"`
}

// CreatedTest is its 201 body.
type CreatedTest struct {
	TestID string `json:"testId"`
}

// BasicInfo is GET/PUT /rest/test-catalog/tests/{testId}/basic-info.
//
// The BOXED flags are the contract, not an accident: a key the caller omits is
// left alone, so a partial PUT cannot silently deactivate a test or clear its
// AMR flag. `name` is read-only here — the Localization section owns it — and
// sending a different one is a 422 rather than a silent no-op.
type BasicInfo struct {
	TestID                  string  `json:"testId,omitempty"`
	Name                    *string `json:"name,omitempty"`
	Code                    *string `json:"code,omitempty"`
	Description             *string `json:"description,omitempty"`
	Domain                  *string `json:"domain,omitempty"`
	LabUnitID               *string `json:"labUnitId,omitempty"`
	SampleTypeID            *string `json:"sampleTypeId,omitempty"`
	AntimicrobialResistance *bool   `json:"antimicrobialResistance,omitempty"`
	Active                  *bool   `json:"active,omitempty"`
	Orderable               *bool   `json:"orderable,omitempty"`
}

// InterpretationDTO is one interpretation rule of a component.
type InterpretationDTO struct {
	ID           string  `json:"id,omitempty"`
	ValueMatch   *string `json:"valueMatch,omitempty"`
	Text         *string `json:"text,omitempty"`
	Severity     *string `json:"severity,omitempty"`
	Color        *string `json:"color,omitempty"`
	DisplayOrder *int    `json:"displayOrder,omitempty"`
}

// OptionDTO is one select-list option. `value` holds the DICTIONARY id — the
// round-trip persists that, and `valueName` is only a label, null when the
// value resolves to no dictionary row.
type OptionDTO struct {
	ID         string  `json:"id,omitempty"`
	Value      *string `json:"value,omitempty"`
	ValueName  *string `json:"valueName,omitempty"`
	ResultType *string `json:"resultType,omitempty"`
	SortOrder  *int    `json:"sortOrder,omitempty"`
	Normal     *bool   `json:"normal,omitempty"`
}

// ResultComponentDTO is one labeled result field of a test.
type ResultComponentDTO struct {
	ID                    string              `json:"id,omitempty"`
	Code                  string              `json:"code,omitempty"`
	Label                 string              `json:"label,omitempty"`
	DisplayOrder          *int                `json:"displayOrder,omitempty"`
	ResultType            *string             `json:"resultType,omitempty"`
	UomID                 *string             `json:"uomId,omitempty"`
	SignificantDigits     *int                `json:"significantDigits,omitempty"`
	DefaultResult         *string             `json:"defaultResult,omitempty"`
	AllowMultipleReadings *bool               `json:"allowMultipleReadings,omitempty"`
	Interpretations       []InterpretationDTO `json:"interpretations"`
	Options               []OptionDTO         `json:"options"`
}

// SampleResults is GET/PUT /rest/test-catalog/tests/{testId}/sample-results,
// and the body of the copy-from shortcut's response.
type SampleResults struct {
	TestID     string               `json:"testId"`
	Components []ResultComponentDTO `json:"components"`
}

// AgeInterval is a half-open age window [fromAge, toAge) in fractional years.
type AgeInterval struct {
	FromAge util.JavaDouble `json:"fromAge"`
	ToAge   util.JavaDouble `json:"toAge"`
}

// SexCoverage is the coverage outcome for one sex.
type SexCoverage struct {
	Sex      string        `json:"sex"`
	Status   string        `json:"status"`
	Gaps     []AgeInterval `json:"gaps"`
	Overlaps []AgeInterval `json:"overlaps"`
}

// CoverageReport is the activation gate's answer — returned on the 409 that
// blocks an activation AND on the 200 that allows one.
type CoverageReport struct {
	Male   *SexCoverage `json:"male,omitempty"`
	Female *SexCoverage `json:"female,omitempty"`
}

// ActivateRequest is the acknowledgment payload: the gap report the operator is
// accepting. Any non-blank string counts as an acknowledgment.
type ActivateRequest struct {
	GapsAcknowledged *string `json:"gapsAcknowledged"`
}

// RangeDTO is one reference range.
//
// Ages are in DAYS, the unit the legacy schema stores — the coverage validator
// then treats them as fractional YEARS, which is a mismatch Java carries and
// the port reproduces. A null bound means unbounded and is serialised from
// ±Infinity, so an unbounded value is ABSENT rather than null in the document.
type RangeDTO struct {
	ID            string           `json:"id,omitempty"`
	ComponentID   *string          `json:"componentId,omitempty"`
	Gender        *string          `json:"gender,omitempty"`
	MinAge        *util.JavaDouble `json:"minAge,omitempty"`
	MaxAge        *util.JavaDouble `json:"maxAge,omitempty"`
	LowNormal     *util.JavaDouble `json:"lowNormal,omitempty"`
	HighNormal    *util.JavaDouble `json:"highNormal,omitempty"`
	LowCritical   *util.JavaDouble `json:"lowCritical,omitempty"`
	HighCritical  *util.JavaDouble `json:"highCritical,omitempty"`
	LowValid      *util.JavaDouble `json:"lowValid,omitempty"`
	HighValid     *util.JavaDouble `json:"highValid,omitempty"`
	LowReporting  *util.JavaDouble `json:"lowReporting,omitempty"`
	HighReporting *util.JavaDouble `json:"highReporting,omitempty"`
}

// RangesResponse is GET/PUT /rest/test-catalog/tests/{testId}/ranges, and the
// PUT's request body. The coverage report is recomputed on every load AND every
// save, so the gap panel always reflects what was just persisted.
type RangesResponse struct {
	TestID   string          `json:"testId"`
	Ranges   []RangeDTO      `json:"ranges"`
	Coverage *CoverageReport `json:"coverage,omitempty"`
}

// GroupRangesUpdate is PUT /rest/test-catalog/group/ranges — the same ranges
// written to every test named.
type GroupRangesUpdate struct {
	TestIDs []string   `json:"testIds"`
	Ranges  []RangeDTO `json:"ranges"`
}
