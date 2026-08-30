package form

// The DTOs the editor's section saves exchange.
//
// Every one of them is a plain public-field Java class serialised with the
// deployment's NON_NULL default, so an unset field is ABSENT from the document
// rather than null — which is why the pointers below are pointers. The blank
// storage document is the clearest case: it carries testId and four false
// booleans and nothing else.

// StorageDTO is GET/PUT /rest/test-catalog/tests/{testId}/storage.
type StorageDTO struct {
	TestID                 string  `json:"testId"`
	StorageCondition       *string `json:"storageCondition,omitempty"`
	StorageConditionCustom *string `json:"storageConditionCustom,omitempty"`
	StorageDuration        *int    `json:"storageDuration,omitempty"`
	StorageDurationUnit    *string `json:"storageDurationUnit,omitempty"`
	StabilityNotes         *string `json:"stabilityNotes,omitempty"`
	ProtectFromLight       *bool   `json:"protectFromLight,omitempty"`
	DoNotFreeze            *bool   `json:"doNotFreeze,omitempty"`
	DoNotRefrigerate       *bool   `json:"doNotRefrigerate,omitempty"`
	DisposalMethod         *string `json:"disposalMethod,omitempty"`
	DisposalTimeframe      *int    `json:"disposalTimeframe,omitempty"`
	DisposalUnit           *string `json:"disposalUnit,omitempty"`
	SpecialInstructions    *string `json:"specialInstructions,omitempty"`
	OverrideRestricted     *bool   `json:"overrideRestricted,omitempty"`
}

// GroupStorageUpdate is PUT /rest/test-catalog/group/storage.
//
// `storage` is applied to every test id in the list, whole — the save is a
// replace, so a field this document omits is CLEARED on each of them.
type GroupStorageUpdate struct {
	TestIDs []string    `json:"testIds"`
	Storage *StorageDTO `json:"storage"`
}

// MappingDTO is one terminology mapping.
type MappingDTO struct {
	ID           string  `json:"id,omitempty"`
	Source       string  `json:"source,omitempty"`
	Code         string  `json:"code,omitempty"`
	Relationship *string `json:"relationship,omitempty"`
}

// TerminologyResponse is GET/PUT /rest/test-catalog/tests/{testId}/terminology.
// The same type is the request body: the PUT reads only `mappings`.
type TerminologyResponse struct {
	TestID   string       `json:"testId"`
	Mappings []MappingDTO `json:"mappings"`
}

// TestOrderRowDTO is one test in a sample type's display order.
type TestOrderRowDTO struct {
	TestID       string  `json:"testId,omitempty"`
	TestName     *string `json:"testName,omitempty"`
	DisplayOrder *int    `json:"displayOrder,omitempty"`
}

// DisplayOrderResponse is GET/PUT
// /rest/test-catalog/sample-types/{sampleTypeId}/test-order.
type DisplayOrderResponse struct {
	SampleTypeID string            `json:"sampleTypeId"`
	Tests        []TestOrderRowDTO `json:"tests"`
}

// TestOrderItem is one entry of the display-order request.
type TestOrderItem struct {
	TestID       string `json:"testId"`
	DisplayOrder *int   `json:"displayOrder"`
}

// DisplayOrderUpdate is the display-order request body.
type DisplayOrderUpdate struct {
	Items []TestOrderItem `json:"items"`
}

// PanelMembershipDTO is one panel a test belongs to.
type PanelMembershipDTO struct {
	PanelID   string  `json:"panelId,omitempty"`
	PanelName *string `json:"panelName,omitempty"`
	Position  *int    `json:"position,omitempty"`
}

// TestPanelsResponse is GET/PUT /rest/test-catalog/tests/{testId}/panels.
type TestPanelsResponse struct {
	TestID      string               `json:"testId"`
	Memberships []PanelMembershipDTO `json:"memberships"`
}

// MembershipItem is one entry of the membership request.
type MembershipItem struct {
	PanelID  string `json:"panelId"`
	Position *int   `json:"position"`
}

// PanelMembershipUpdate is the membership request body.
type PanelMembershipUpdate struct {
	Memberships []MembershipItem `json:"memberships"`
}

// PanelTestRowDTO is one test inside a panel.
type PanelTestRowDTO struct {
	TestID   string  `json:"testId,omitempty"`
	TestName *string `json:"testName,omitempty"`
	Position *int    `json:"position,omitempty"`
}

// PanelTestOrderResponse is GET /rest/test-catalog/panels/{panelId}/test-order.
type PanelTestOrderResponse struct {
	PanelID string            `json:"panelId"`
	Tests   []PanelTestRowDTO `json:"tests"`
}

// CreatePanelRequest is POST /rest/test-catalog/panels.
type CreatePanelRequest struct {
	Name string `json:"name"`
}
