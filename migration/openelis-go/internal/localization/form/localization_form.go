// Package form carries the localization DTOs.
package form

import "sort"

// LocalizationDTO is a fully serialised org.openelisglobal.localization.valueholder.Localization.
//
// Jackson serialises every public getter, and Localization has nine derived
// ones on top of its three stored fields — so a single nested localization
// object ships the same three values in seven different shapes. None of it is
// selected by an annotation; it is simply what the bean exposes.
//
// Field order is the wire contract: stored fields first, then the getters in
// declaration order.
type LocalizationDTO struct {
	Lastupdated *int64                      `json:"lastupdated,omitempty"`
	ID          string                      `json:"id"`
	Description string                      `json:"description"`
	Values      map[string]LocalizationValue `json:"values"`

	// LocalizedValue is the value for the CURRENT request locale, falling back
	// to English.
	LocalizedValue   string   `json:"localizedValue"`
	LocalesWithValue []string `json:"localesWithValue"`

	// English and French are @Deprecated getters that Jackson serialises
	// anyway — two more copies of what `values` already holds.
	English string `json:"english"`
	French  string `json:"french"`

	ValuesAsMap map[string]string `json:"valuesAsMap"`

	AllActiveLocales                 []string `json:"allActiveLocales"`
	LocalesWithValueSortedForDisplay []string `json:"localesWithValueSortedForDisplay"`
	LocalesSortedForDisplay          []string `json:"localesSortedForDisplay"`

	// LocalesAndValuesOfLocalesWithValues is "<display language>: <value>", and
	// the display language comes from the JDK — Locale.getDisplayLanguage — not
	// from supported_locale.display_name. The two differ on this data: the
	// column says "Francais" and Java answers "French".
	LocalesAndValuesOfLocalesWithValues []string `json:"localesAndValuesOfLocalesWithValues"`
}

// LocalizationValue is one entry of the `values` map.
type LocalizationValue struct {
	Lastupdated *int64 `json:"lastupdated,omitempty"`
	ID          string `json:"id"`
	Locale      string `json:"locale"`
	Value       string `json:"value"`
}

// displayLanguage is Locale.getDisplayLanguage(Locale.ENGLISH) for the locales
// this deployment activates.
//
// The JDK knows every language tag; this map knows two. A locale outside it
// falls back to its own code, which is wrong but visible — the alternative,
// reaching for supported_locale.display_name, would be wrong and invisible,
// because that column reads "Francais" where Java answers "French".
var displayLanguage = map[string]string{
	"en": "English",
	"fr": "French",
}

// DisplayLanguage resolves a locale code to the name Java renders.
func DisplayLanguage(code string) string {
	if name, ok := displayLanguage[code]; ok {
		return name
	}
	return code
}

// BuildLocalization assembles the DTO from the stored row plus the active
// locale list.
//
// requestLocale is LocaleContextHolder.getLocale() — the locale localizedValue
// and the display names are resolved against.
func BuildLocalization(id string, description string, lastupdated *int64,
	values []LocalizationValue, activeLocales []string, requestLocale string) *LocalizationDTO {

	byLocale := map[string]LocalizationValue{}
	for _, v := range values {
		byLocale[v.Locale] = v
	}

	// getLocalesWithValue skips a blank value. Sorted by locale code: Java
	// iterates a HashMap here, and for the two keys this deployment holds the
	// bucket order and the alphabetical order agree.
	withValue := []string{}
	asMap := map[string]string{}
	for _, v := range values {
		asMap[v.Locale] = v.Value
		if v.Value != "" {
			withValue = append(withValue, v.Locale)
		}
	}
	sort.Strings(withValue)

	// sortLocales compares the DISPLAY language, not the code.
	sortByDisplay := func(codes []string) []string {
		out := append([]string(nil), codes...)
		sort.SliceStable(out, func(i, j int) bool {
			return DisplayLanguage(out[i]) < DisplayLanguage(out[j])
		})
		return out
	}

	withValueSorted := sortByDisplay(withValue)
	pairs := make([]string, 0, len(withValueSorted))
	for _, code := range withValueSorted {
		pairs = append(pairs, DisplayLanguage(code)+": "+localized(byLocale, code))
	}

	return &LocalizationDTO{
		Lastupdated:                         lastupdated,
		ID:                                  id,
		Description:                         description,
		Values:                              byLocale,
		LocalizedValue:                      localized(byLocale, requestLocale),
		LocalesWithValue:                    withValue,
		English:                             byLocale["en"].Value,
		French:                              byLocale["fr"].Value,
		ValuesAsMap:                         asMap,
		AllActiveLocales:                    activeLocales,
		LocalesWithValueSortedForDisplay:    withValueSorted,
		LocalesSortedForDisplay:             sortByDisplay(activeLocales),
		LocalesAndValuesOfLocalesWithValues: pairs,
	}
}

// localized ports getLocalizedValue(Locale): the exact locale's value when it
// is non-blank, else the English fallback, else "".
func localized(byLocale map[string]LocalizationValue, code string) string {
	if v, ok := byLocale[code]; ok && v.Value != "" {
		return v.Value
	}
	if v, ok := byLocale["en"]; ok && v.Value != "" {
		return v.Value
	}
	return ""
}
