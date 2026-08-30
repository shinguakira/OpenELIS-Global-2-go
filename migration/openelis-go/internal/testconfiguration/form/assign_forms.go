package form

import "openelis-go/internal/common/util"

// SelectedPanelDTO is PanelTestAssign's `selectedPanel`: the tests currently in
// the panel and the ones that could be added.
type SelectedPanelDTO struct {
	Tests          []util.IdValuePair `json:"tests"`
	AvailableTests []util.IdValuePair `json:"availableTests"`
}

// AssignForm is GET/POST for the three *TestAssign screens.
//
// One type for three beans; the pointers keep the other screens' fields out of
// the document. The id fields are initialised to "" on every bean, so they are
// present on the blank form.
type AssignForm struct {
	FormName       string `json:"formName"`
	FormMethod     string `json:"formMethod"`
	CancelAction   string `json:"cancelAction"`
	SubmitOnCancel bool   `json:"submitOnCancel"`
	CancelMethod   string `json:"cancelMethod"`

	PanelList       *[]util.IdValuePair `json:"panelList,omitempty"`
	SampleTypeList  *[]util.IdValuePair `json:"sampleTypeList,omitempty"`
	TestSectionList *[]util.IdValuePair `json:"testSectionList,omitempty"`

	// SampleTypeTestList and SectionTestList are maps whose KEYS are the
	// IdValuePair.toString() of the sample type or section —
	// `"id=31, value=Sputum"` — because the bean declares a
	// LinkedHashMap<IdValuePair, List<IdValuePair>> and Jackson renders a
	// non-String key with toString(). The blank leading entry is there too, as
	// `"id=0, value="` mapped to an empty list.
	SampleTypeTestList *map[string][]util.IdValuePair `json:"sampleTypeTestList,omitempty"`
	SectionTestList    *map[string][]util.IdValuePair `json:"sectionTestList,omitempty"`

	SelectedPanel *SelectedPanelDTO `json:"selectedPanel,omitempty"`

	TestID                  *string `json:"testId,omitempty"`
	SampleTypeID            *string `json:"sampleTypeId,omitempty"`
	DeactivateSampleTypeID  *string `json:"deactivateSampleTypeId,omitempty"`
	TestSectionID           *string `json:"testSectionId,omitempty"`
	DeactivateTestSectionID *string `json:"deactivateTestSectionId,omitempty"`
	PanelID                 *string `json:"panelId,omitempty"`
	DeactivatePanelID       *string `json:"deactivatePanelId,omitempty"`
}

// AssignPost is the bound body for all three.
type AssignPost struct {
	TestID                  *string  `json:"testId"`
	SampleTypeID            *string  `json:"sampleTypeId"`
	DeactivateSampleTypeID  *string  `json:"deactivateSampleTypeId"`
	TestSectionID           *string  `json:"testSectionId"`
	DeactivateTestSectionID *string  `json:"deactivateTestSectionId"`
	PanelID                 *string  `json:"panelId"`
	DeactivatePanelID       *string  `json:"deactivatePanelId"`
	CurrentTests            []string `json:"currentTests"`
}

// NewAssignForm reproduces `new XTestAssignForm()`.
func NewAssignForm(formName string) AssignForm {
	empty := ""
	return AssignForm{
		FormName: formName, FormMethod: "POST",
		CancelAction: "Home", SubmitOnCancel: false, CancelMethod: "POST",
		TestID: &empty,
	}
}
