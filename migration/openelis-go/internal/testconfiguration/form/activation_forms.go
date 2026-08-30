package form

import "openelis-go/internal/common/util"

// TestActivationBean is one sample type and the tests under it, split two ways.
//
// TestActivation splits on is_active; TestOrderability splits on orderable and
// reuses the same bean, so the keys are named for the first screen either way.
type TestActivationBean struct {
	SampleType    util.IdValuePair   `json:"sampleType"`
	ActiveTests   []util.IdValuePair `json:"activeTests"`
	InactiveTests []util.IdValuePair `json:"inactiveTests"`
}

// ActivationForm is GET/POST for TestActivation and TestOrderability.
//
// One type for two beans; the pointers keep the other screen's list out of the
// document. jsonChangeList is initialised to "" on both, so it is present on
// the blank form.
type ActivationForm struct {
	FormName       string `json:"formName"`
	FormMethod     string `json:"formMethod"`
	CancelAction   string `json:"cancelAction"`
	SubmitOnCancel bool   `json:"submitOnCancel"`
	CancelMethod   string `json:"cancelMethod"`

	ActiveTestList    *[]TestActivationBean `json:"activeTestList,omitempty"`
	InactiveTestList  *[]TestActivationBean `json:"inactiveTestList,omitempty"`
	OrderableTestList *[]TestActivationBean `json:"orderableTestList,omitempty"`

	JSONChangeList string `json:"jsonChangeList"`
}

// ActivationPost is the bound body — the same double-encoded change list the
// *Order screens take, with four keys rather than one.
type ActivationPost struct {
	JSONChangeList *string `json:"jsonChangeList"`
}

// NewActivationForm reproduces `new TestActivationForm()` /
// `new TestOrderabilityForm()`.
func NewActivationForm(formName string) ActivationForm {
	return ActivationForm{
		FormName: formName, FormMethod: "POST",
		CancelAction: "Home", SubmitOnCancel: false, CancelMethod: "POST",
	}
}
