package form

import "openelis-go/internal/common/util"

// The *RenameEntry forms.
//
// They share a shape — a list of what can be renamed, then the three or five
// bound fields — but each names its list and its id field after its own entity,
// so they are separate types rather than one generic form with a renamed key.
// Field ORDER is Jackson's: BaseForm first, then the bean's own declarations.
//
// nameEnglish, nameFrench and the id are initialised to "" on the Java beans
// rather than left null, so they are PRESENT on the blank form as empty
// strings. The list is a pointer because it is absent from the POST response:
// only the validation-failure branch re-populates it.

// MethodRenameEntryForm is GET/POST /rest/MethodRenameEntry.
type MethodRenameEntryForm struct {
	FormName       string `json:"formName"`
	FormMethod     string `json:"formMethod"`
	CancelAction   string `json:"cancelAction"`
	SubmitOnCancel bool   `json:"submitOnCancel"`
	CancelMethod   string `json:"cancelMethod"`

	MethodList  *[]util.IdValuePair `json:"methodList,omitempty"`
	NameEnglish string              `json:"nameEnglish"`
	NameFrench  string              `json:"nameFrench"`
	MethodID    string              `json:"methodId"`
}

// PanelRenameEntryForm is GET/POST /rest/PanelRenameEntry.
type PanelRenameEntryForm struct {
	FormName       string `json:"formName"`
	FormMethod     string `json:"formMethod"`
	CancelAction   string `json:"cancelAction"`
	SubmitOnCancel bool   `json:"submitOnCancel"`
	CancelMethod   string `json:"cancelMethod"`

	PanelList   *[]util.IdValuePair `json:"panelList,omitempty"`
	NameEnglish string              `json:"nameEnglish"`
	NameFrench  string              `json:"nameFrench"`
	PanelID     string              `json:"panelId"`
}

// SampleTypeRenameEntryForm is GET/POST /rest/SampleTypeRenameEntry.
type SampleTypeRenameEntryForm struct {
	FormName       string `json:"formName"`
	FormMethod     string `json:"formMethod"`
	CancelAction   string `json:"cancelAction"`
	SubmitOnCancel bool   `json:"submitOnCancel"`
	CancelMethod   string `json:"cancelMethod"`

	SampleTypeList *[]util.IdValuePair `json:"sampleTypeList,omitempty"`
	NameEnglish    string              `json:"nameEnglish"`
	NameFrench     string              `json:"nameFrench"`
	SampleTypeID   string              `json:"sampleTypeId"`
}

// TestSectionRenameEntryForm is GET/POST /rest/TestSectionRenameEntry.
type TestSectionRenameEntryForm struct {
	FormName       string `json:"formName"`
	FormMethod     string `json:"formMethod"`
	CancelAction   string `json:"cancelAction"`
	SubmitOnCancel bool   `json:"submitOnCancel"`
	CancelMethod   string `json:"cancelMethod"`

	TestSectionList *[]util.IdValuePair `json:"testSectionList,omitempty"`
	NameEnglish     string              `json:"nameEnglish"`
	NameFrench      string              `json:"nameFrench"`
	TestSectionID   string              `json:"testSectionId"`
}

// RenamePost is the bound body for the four screens above. ALLOWED_FIELDS
// differs per controller only in the name of the id field, so each is decoded
// into its own struct and normalised by the service.
type RenamePost struct {
	MethodID      *string `json:"methodId"`
	PanelID       *string `json:"panelId"`
	SampleTypeID  *string `json:"sampleTypeId"`
	TestSectionID *string `json:"testSectionId"`
	NameEnglish   *string `json:"nameEnglish"`
	NameFrench    *string `json:"nameFrench"`
}

// NewMethodRenameEntryForm reproduces `new MethodRenameEntryForm()`.
func NewMethodRenameEntryForm() MethodRenameEntryForm {
	return MethodRenameEntryForm{
		FormName: "methodRenameEntryForm", FormMethod: "POST",
		CancelAction: "Home", SubmitOnCancel: false, CancelMethod: "POST",
	}
}

// NewPanelRenameEntryForm reproduces `new PanelRenameEntryForm()`.
func NewPanelRenameEntryForm() PanelRenameEntryForm {
	return PanelRenameEntryForm{
		FormName: "panelRenameEntryForm", FormMethod: "POST",
		CancelAction: "Home", SubmitOnCancel: false, CancelMethod: "POST",
	}
}

// NewSampleTypeRenameEntryForm reproduces `new SampleTypeRenameEntryForm()`.
func NewSampleTypeRenameEntryForm() SampleTypeRenameEntryForm {
	return SampleTypeRenameEntryForm{
		FormName: "sampleTypeRenameEntryForm", FormMethod: "POST",
		CancelAction: "Home", SubmitOnCancel: false, CancelMethod: "POST",
	}
}

// NewTestSectionRenameEntryForm reproduces `new TestSectionRenameEntryForm()`.
func NewTestSectionRenameEntryForm() TestSectionRenameEntryForm {
	return TestSectionRenameEntryForm{
		FormName: "testSectionRenameEntryForm", FormMethod: "POST",
		CancelAction: "Home", SubmitOnCancel: false, CancelMethod: "POST",
	}
}
