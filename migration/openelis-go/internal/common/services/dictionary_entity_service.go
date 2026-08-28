package services

import (
	"time"

	"openelis-go/internal/common/daoimpl"
	"openelis-go/internal/common/form"
)

// activeLocale is the locale the form loads render in. Java resolves it from
// the request; every response measured here is English, and the suite logs in
// with the English default.
const activeLocale = "en"

// millis renders a timestamp the way Jackson writes java.sql.Timestamp: epoch
// milliseconds, not ISO-8601. 0 for a null column.
func millis(t *time.Time) int64 {
	if t == nil {
		return 0
	}
	return t.UnixMilli()
}

// DictionaryEntities builds the FULL entity form of a dictionary category —
// the shape patientProperties.addressDepartments and
// projectDataEID/VL.hivStatusList carry, as opposed to the {id,value} pairs the
// sibling lists use.
func (s *DisplayListService) DictionaryEntities(categoryName string) ([]form.DictionaryEntityDTO, error) {
	rows, err := s.DAO.DictionaryEntitiesByCategory(categoryName)
	if err != nil {
		return nil, err
	}

	locales, err := s.DAO.ActiveSupportedLocales()
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(rows))
	for _, r := range rows {
		if r.LocalizationID != nil {
			ids = append(ids, *r.LocalizationID)
		}
	}
	locs, vals, err := s.DAO.LocalizationsByIDs(ids)
	if err != nil {
		return nil, err
	}

	out := make([]form.DictionaryEntityDTO, 0, len(rows))
	for _, r := range rows {
		dto := form.DictionaryEntityDTO{
			Lastupdated:       millis(r.Lastupdated),
			ID:                r.ID,
			IsActive:          r.IsActive,
			DictEntry:         r.DictEntry,
			LocalAbbreviation: r.LocalAbbrev,
			NameKey:           r.DisplayKey,
			DictionaryCategory: form.DictionaryCategoryDTO{
				Lastupdated:       millis(r.CategoryLastupdated),
				ID:                r.CategoryID,
				Description:       r.CategoryDescription,
				LocalAbbreviation: r.CategoryLocalAbbrev,
				CategoryName:      r.CategoryName,
			},
		}
		if r.SortOrder != nil {
			dto.SortOrder = *r.SortOrder
		}

		if r.LocalizationID != nil {
			loc := locs[*r.LocalizationID]
			dto.LocalizedDictionaryName = buildLocalization(loc, vals[*r.LocalizationID], locales)
		}

		// getLocalizedName falls back to dict_entry when the locale has no
		// value — the same fallback the {id,value} lists use.
		dto.DisplayValue = dto.LocalizedDictionaryName.LocalizedValue
		if dto.DisplayValue == "" {
			dto.DisplayValue = r.DictEntry
		}
		out = append(out, dto)
	}
	return out, nil
}

// buildLocalization derives every computed member of the Localization entity.
// Only id, description and lastupdated are stored; the rest are getters, and
// Jackson serializes getters, so all of them reach the wire.
func buildLocalization(loc daoimpl.LocalizationRow, values []daoimpl.LocalizationValueRow,
	locales []daoimpl.SupportedLocaleRow) form.LocalizationDTO {

	byLocale := map[string]daoimpl.LocalizationValueRow{}
	for _, v := range values {
		byLocale[v.Locale] = v
	}

	dto := form.LocalizationDTO{
		Lastupdated: millis(loc.Lastupdated),
		ID:          loc.ID,
		Description: loc.Description,
		Values:      map[string]form.LocalizationValueDTO{},
		ValuesAsMap: map[string]string{},
		// Non-nil so every list serializes as [] rather than null when empty.
		LocalesWithValue:                    []string{},
		AllActiveLocales:                    []string{},
		LocalesWithValueSortedForDisplay:    []string{},
		LocalesSortedForDisplay:             []string{},
		LocalesAndValuesOfLocalesWithValues: []string{},
	}

	for _, v := range values {
		dto.Values[v.Locale] = form.LocalizationValueDTO{
			Lastupdated: millis(v.LastUpdated),
			ID:          v.ID,
			Locale:      v.Locale,
			Value:       v.Value,
		}
		dto.ValuesAsMap[v.Locale] = v.Value
		dto.LocalesWithValue = append(dto.LocalesWithValue, v.Locale)
	}

	// The display-ordered lists follow supported_locale.sort_order, so they are
	// built by walking the locales rather than the values.
	for _, l := range locales {
		dto.AllActiveLocales = append(dto.AllActiveLocales, l.LocaleCode)
		dto.LocalesSortedForDisplay = append(dto.LocalesSortedForDisplay, l.LocaleCode)
		if v, ok := byLocale[l.LocaleCode]; ok {
			dto.LocalesWithValueSortedForDisplay = append(dto.LocalesWithValueSortedForDisplay, l.LocaleCode)
			// "<display name>: <value>" — the locale's human label, not its code.
			dto.LocalesAndValuesOfLocalesWithValues = append(
				dto.LocalesAndValuesOfLocalesWithValues, l.DisplayName+": "+v.Value)
		}
	}

	dto.LocalizedValue = byLocale[activeLocale].Value
	dto.English = byLocale["en"].Value
	dto.French = byLocale["fr"].Value
	return dto
}

// PatientTypeEntities builds the full patient_type entities.
func (s *DisplayListService) PatientTypeEntities() ([]form.PatientTypeEntityDTO, error) {
	rows, err := s.DAO.PatientTypes()
	if err != nil {
		return nil, err
	}
	out := make([]form.PatientTypeEntityDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, form.PatientTypeEntityDTO{
			Lastupdated: millis(r.Lastupdated),
			// Not a column — the base class initialises it to "Y".
			IsActive:    "Y",
			ID:          r.ID,
			Type:        r.Type,
			Description: r.Description,
		})
	}
	return out, nil
}

// ProjectEntities builds the full project entities.
func (s *DisplayListService) ProjectEntities() ([]form.ProjectEntityDTO, error) {
	rows, err := s.DAO.Projects()
	if err != nil {
		return nil, err
	}
	out := make([]form.ProjectEntityDTO, 0, len(rows))
	for _, r := range rows {
		// concatProjNameDesc is derived: name + "+" + description, and just the
		// name when there is no description.
		concat := r.ProjectName
		if r.Description != nil {
			concat = r.ProjectName + "+" + *r.Description
		}
		out = append(out, form.ProjectEntityDTO{
			Lastupdated:        millis(r.Lastupdated),
			ID:                 r.ID,
			ProjectName:        r.ProjectName,
			Description:        r.Description,
			IsActive:           r.IsActive,
			ProgramCode:        r.ProgramCode,
			ConcatProjNameDesc: concat,
			// Never populated on these form loads, but initialised on the
			// entity, so it serializes as [].
			Organizations: []any{},
		})
	}
	return out, nil
}
