package form

import "openelis-go/internal/common/util"

// UomCreateForm is what GET and POST /rest/UomCreate answer — and they are not
// the same document.
//
// Field ORDER is the wire contract: Jackson serialises a bean in declaration
// order with BaseForm's fields first, so formName…cancelMethod precede the
// UomCreateForm fields.
//
// The four list/name fields are POINTERS because their PRESENCE is observable.
// setupDisplayItems fills them, and the POST success branch does not call it:
// a create answers formName, formMethod, cancelAction, submitOnCancel,
// cancelMethod and uomEnglishName, with the lists absent under
// Include.NON_NULL. The validation-failure branch does call it — and also
// answers 200 — so the two 200s carry different key sets.
type UomCreateForm struct {
	FormName       string `json:"formName"`
	FormMethod     string `json:"formMethod"`
	CancelAction   string `json:"cancelAction"`
	SubmitOnCancel bool   `json:"submitOnCancel"`
	CancelMethod   string `json:"cancelMethod"`

	ExistingUomList *[]util.IdValuePair `json:"existingUomList,omitempty"`
	InactiveUomList *[]util.IdValuePair `json:"inactiveUomList,omitempty"`
	// ExistingEnglishNames is "$name1$name2$" — getExistingUomNames seeds the
	// builder with the separator and appends one after every name, so the value
	// carries a leading AND a trailing "$".
	ExistingEnglishNames *string `json:"existingEnglishNames,omitempty"`
	// ExistingFrenchNames is the same shape and is NOT French. It is the
	// literal word "French", once per UOM: unit_of_measure has no localization
	// column, and UnitOfMeasure.getLocalization() is a stub that builds a
	// Localization in memory per call, ending with setFrench("French"). The
	// comment above it says UOM "has been designed to support localization" and
	// that the columns were never added. Reproduced, not repaired.
	ExistingFrenchNames *string `json:"existingFrenchNames,omitempty"`

	UomEnglishName *string `json:"uomEnglishName,omitempty"`
	UomFrenchName  *string `json:"uomFrenchName,omitempty"`
}

// NewUomCreateForm reproduces `new UomCreateForm()` — BaseForm's initialisers
// plus the constructor's setFormName.
func NewUomCreateForm() UomCreateForm {
	return UomCreateForm{
		FormName:       "uomCreateForm",
		FormMethod:     "POST",
		CancelAction:   "Home",
		SubmitOnCancel: false,
		CancelMethod:   "POST",
	}
}

// UomCreatePost is the request body. Only uomEnglishName is bound:
// initBinder's ALLOWED_FIELDS is {"uomEnglishName"}, so uomFrenchName is
// declared on the bean, accepted by the parser and never applied.
type UomCreatePost struct {
	UomEnglishName *string `json:"uomEnglishName"`
}

// UomRenameEntryForm is GET/POST /rest/UomRenameEntry.
//
// nameEnglish, nameFrench and uomId are initialised to "" on the bean rather
// than left null, so they are present on every response — including the blank
// form, where they are empty strings and not absent keys.
type UomRenameEntryForm struct {
	FormName       string `json:"formName"`
	FormMethod     string `json:"formMethod"`
	CancelAction   string `json:"cancelAction"`
	SubmitOnCancel bool   `json:"submitOnCancel"`
	CancelMethod   string `json:"cancelMethod"`

	UomList *[]util.IdValuePair `json:"uomList,omitempty"`

	NameEnglish string `json:"nameEnglish"`
	NameFrench  string `json:"nameFrench"`
	UomID       string `json:"uomId"`
}

// NewUomRenameEntryForm reproduces `new UomRenameEntryForm()`.
func NewUomRenameEntryForm() UomRenameEntryForm {
	return UomRenameEntryForm{
		FormName:       "uomRenameEntryForm",
		FormMethod:     "POST",
		CancelAction:   "Home",
		SubmitOnCancel: false,
		CancelMethod:   "POST",
	}
}

// UomRenameEntryPost is the request body — ALLOWED_FIELDS is
// {"uomId", "nameEnglish"}, so nameFrench is on the bean and never bound.
type UomRenameEntryPost struct {
	UomID       *string `json:"uomId"`
	NameEnglish *string `json:"nameEnglish"`
}
