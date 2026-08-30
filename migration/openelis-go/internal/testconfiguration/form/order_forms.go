package form

import "openelis-go/internal/common/util"

// OrderFormNames maps each *Order screen to the formName its bean reports.
var OrderFormNames = map[string]string{
	"PanelOrder":       "panelOrderForm",
	"SampleTypeOrder":  "sampleTypeOrderForm",
	"TestSectionOrder": "testSectionOrderForm",
}

// OrderForm is GET/POST for the three *Order screens.
//
// One type for three beans: they differ only in which list they carry, and the
// pointers keep the other two out of the document under Include.NON_NULL.
// jsonChangeList is initialised to "" on every one of them, so it is present on
// the blank form as an empty string rather than an absent key.
type OrderForm struct {
	FormName       string `json:"formName"`
	FormMethod     string `json:"formMethod"`
	CancelAction   string `json:"cancelAction"`
	SubmitOnCancel bool   `json:"submitOnCancel"`
	CancelMethod   string `json:"cancelMethod"`

	PanelList       *[]util.IdValuePair `json:"panelList,omitempty"`
	SampleTypeList  *[]util.IdValuePair `json:"sampleTypeList,omitempty"`
	TestSectionList *[]util.IdValuePair `json:"testSectionList,omitempty"`

	JSONChangeList string `json:"jsonChangeList"`

	// PanelOrder's bean also carries the PanelCreate lists, because it extends
	// that form. They are absent unless the screen fills them.
	ExistingPanelList      *[]SampleTypePanelDTO `json:"existingPanelList,omitempty"`
	InactivePanelList      *[]SampleTypePanelDTO `json:"inactivePanelList,omitempty"`
	ExistingSampleTypeList *[]util.IdValuePair   `json:"existingSampleTypeList,omitempty"`
}

// OrderPost is the bound body — one field, and it is a JSON STRING rather than
// a nested object. See parseChangeList.
type OrderPost struct {
	JSONChangeList *string `json:"jsonChangeList"`
}

// NewOrderForm reproduces `new XOrderForm()`.
func NewOrderForm(formName string) OrderForm {
	return OrderForm{
		FormName: formName, FormMethod: "POST",
		CancelAction: "Home", SubmitOnCancel: false, CancelMethod: "POST",
	}
}
