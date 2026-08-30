package form

import "openelis-go/internal/common/util"

// TestAddForm is GET/POST /rest/TestAdd.
//
// The GET fills eight lists and leaves jsonWad at "". `loinc` is a declared
// field but the GET sets it from `new Test().getLoinc()`, which is null — and
// the form is NON_NULL, so the KEY IS ABSENT on the blank form. The POST echoes
// whatever was bound, so it is present there.
//
// The lists are pointers for the same reason: the POST re-serialises the bound
// form without refilling them, so all eight keys vanish from the response to a
// create.
type TestAddForm struct {
	FormName       string `json:"formName"`
	FormMethod     string `json:"formMethod"`
	CancelAction   string `json:"cancelAction"`
	SubmitOnCancel bool   `json:"submitOnCancel"`
	CancelMethod   string `json:"cancelMethod"`

	JSONWad string  `json:"jsonWad"`
	Loinc   *string `json:"loinc,omitempty"`

	SampleTypeList        *[]util.IdValuePair   `json:"sampleTypeList,omitempty"`
	PanelList             *[]util.IdValuePair   `json:"panelList,omitempty"`
	UomList               *[]util.IdValuePair   `json:"uomList,omitempty"`
	ResultTypeList        *[]util.IdValuePair   `json:"resultTypeList,omitempty"`
	AgeRangeList          *[]util.IdValuePair   `json:"ageRangeList,omitempty"`
	LabUnitList           *[]util.IdValuePair   `json:"labUnitList,omitempty"`
	DictionaryList        *[]util.IdValuePair   `json:"dictionaryList,omitempty"`
	GroupedDictionaryList *[][]util.IdValuePair `json:"groupedDictionaryList,omitempty"`
}

// TestAddPost is the bound body. jsonWad carries the whole submission as a
// STRING of JSON — every field the screen collects is inside it, and none of
// them is a form property.
type TestAddPost struct {
	JSONWad *string `json:"jsonWad"`
	Loinc   *string `json:"loinc"`
}

// NewTestAddForm reproduces `new TestAddForm()`.
func NewTestAddForm() TestAddForm {
	return TestAddForm{
		FormName: "testAddForm", FormMethod: "POST",
		CancelAction: "Home", SubmitOnCancel: false, CancelMethod: "POST",
		JSONWad: "",
	}
}
