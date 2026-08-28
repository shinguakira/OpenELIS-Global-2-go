// Package form ports the shared entity shapes the Wave 4 form loads serialize.
// Folder layout mirrors the Java source during migration.
package form

// DictionaryEntityDTO is a full org.openelisglobal.dictionary.valueholder.Dictionary
// as Jackson writes it — NOT the {id,value} pair the DisplayListService lists use.
//
// Several form-load fields carry this heavier shape (patientProperties
// .addressDepartments, projectDataEID/VL.hivStatusList) while their siblings in
// the same object carry {id,value}. A port that emits pairs everywhere returns
// the right ids and the wrong response.
//
// NameKey and LocalAbbreviation are pointers because Include.NON_NULL drops
// them per row: a dictionary entry with no display_key omits `nameKey`
// entirely, and one with a NULL local_abbrev omits `localAbbreviation` while
// one storing "" emits it as "". Both cases occur in the dev dataset.
type DictionaryEntityDTO struct {
	// Epoch millis — Jackson's default for java.sql.Timestamp.
	Lastupdated int64 `json:"lastupdated"`
	// Present only when display_key is set.
	NameKey *string `json:"nameKey,omitempty"`

	ID       string `json:"id"`
	IsActive string `json:"isActive"`

	DictEntry          string                `json:"dictEntry"`
	DictionaryCategory DictionaryCategoryDTO `json:"dictionaryCategory"`
	LocalAbbreviation  *string               `json:"localAbbreviation,omitempty"`
	SortOrder          int64                 `json:"sortOrder"`

	LocalizedDictionaryName LocalizationDTO `json:"localizedDictionaryName"`
	// getLocalizedName() for the active locale, falling back to dict_entry.
	DisplayValue string `json:"displayValue"`
}

// DictionaryCategoryDTO is the nested category. Note `categoryName` is the DB
// column `name` and `localAbbreviation` is `local_abbrev` — neither JSON key
// matches its column.
type DictionaryCategoryDTO struct {
	Lastupdated       int64  `json:"lastupdated"`
	ID                string `json:"id"`
	Description       string `json:"description"`
	LocalAbbreviation string `json:"localAbbreviation"`
	CategoryName      string `json:"categoryName"`
}

// LocalizationDTO is org.openelisglobal.localization.valueholder.Localization.
//
// Most of this object is DERIVED, not stored: clinlims.localization holds only
// (id, description, lastupdated), and the per-locale text lives in
// clinlims.localization_value. Every list below is computed by a getter, and
// Jackson serializes getters, so all of them appear on the wire:
//
//	values                              map keyed by locale -> the full value row
//	valuesAsMap                         the same, flattened to locale -> string
//	localizedValue                      the active locale's value
//	english / french                    named accessors, "" when absent
//	localesWithValue                    locales that actually have a row
//	allActiveLocales                    every active supported_locale
//	localesSortedForDisplay             those, in supported_locale.sort_order
//	localesWithValueSortedForDisplay    the intersection, same order
//	localesAndValuesOfLocalesWithValues "<display_name>: <value>" per locale
//
// A port that stores only the resolved string cannot produce any of it.
type LocalizationDTO struct {
	Lastupdated int64                           `json:"lastupdated"`
	ID          string                          `json:"id"`
	Description string                          `json:"description"`
	Values      map[string]LocalizationValueDTO `json:"values"`

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

// LocalizationValueDTO is one clinlims.localization_value row.
type LocalizationValueDTO struct {
	Lastupdated int64  `json:"lastupdated"`
	ID          string `json:"id"`
	Locale      string `json:"locale"`
	Value       string `json:"value"`
}

// PatientTypeEntityDTO is a full patient_type row.
//
// `isActive` is NOT a column — patient_type has none. It comes from
// EnumValueItemImpl, whose field is initialised to the literal "Y" and is never
// overwritten on this path, so it is a constant in the response.
type PatientTypeEntityDTO struct {
	Lastupdated int64  `json:"lastupdated"`
	IsActive    string `json:"isActive"`
	ID          string `json:"id"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// ProjectEntityDTO is a full clinlims.project row.
//
// `concatProjNameDesc` is DERIVED (name + "+" + description), not stored, and
// `projectName` is the column `name`. ProgramCode is nullable and dropped by
// NON_NULL when absent.
type ProjectEntityDTO struct {
	Lastupdated        int64   `json:"lastupdated"`
	ID                 string  `json:"id"`
	ProjectName        string  `json:"projectName"`
	Description        *string `json:"description,omitempty"`
	IsActive           string  `json:"isActive"`
	ProgramCode        *string `json:"programCode,omitempty"`
	ConcatProjNameDesc string  `json:"concatProjNameDesc"`
	Organizations      []any   `json:"organizations"`
}
