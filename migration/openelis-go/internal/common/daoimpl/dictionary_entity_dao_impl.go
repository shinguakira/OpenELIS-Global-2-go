package daoimpl

import (
	"time"
)

// DictionaryEntityRow is one dictionary row joined to its category, flat.
// The localization half is fetched separately because it is one-to-many.
type DictionaryEntityRow struct {
	ID          string     `gorm:"column:id"`
	IsActive    string     `gorm:"column:is_active"`
	DictEntry   string     `gorm:"column:dict_entry"`
	LocalAbbrev *string    `gorm:"column:local_abbrev"`
	DisplayKey  *string    `gorm:"column:display_key"`
	SortOrder   *int64     `gorm:"column:sort_order"`
	Lastupdated *time.Time `gorm:"column:lastupdated"`

	LocalizationID *string `gorm:"column:name_localization_id"`

	CategoryID          string     `gorm:"column:category_id"`
	CategoryDescription string     `gorm:"column:category_description"`
	CategoryLocalAbbrev string     `gorm:"column:category_local_abbrev"`
	CategoryName        string     `gorm:"column:category_name"`
	CategoryLastupdated *time.Time `gorm:"column:category_lastupdated"`
}

// DictionaryEntitiesByCategory returns full dictionary entities for a category.
//
// Two filters that are easy to miss, both measured against the live response:
//   - is_active = 'Y'. getDictionaryEntrysByCategoryAbbreviation hardcodes it in
//     every branch of its HQL, so an inactive entry never reaches the form.
//   - the order is the LOCALIZED NAME, not sort_order — BaseObject
//     .sortByLocalizedName runs after the query. addressDepartments therefore
//     starts at "Artibonite" (sort_order 86300) rather than "Ouest" (86000).
//
// The category is matched on dictionary_category.name, and TWO categories can
// share a name: `HIVResult` exists twice (local_abbrev `HIVResult` with 4
// entries and `Conclusion` with 3). Java matches on the name alone and returns
// all seven, interleaved by the localized sort — so filtering to one category
// would return the wrong list.
func (d *DisplayListDAOImpl) DictionaryEntitiesByCategory(categoryName string) ([]DictionaryEntityRow, error) {
	rows := []DictionaryEntityRow{}
	err := d.DB.Table("clinlims.dictionary AS d").
		Select(`d.id, d.is_active, d.dict_entry, d.local_abbrev, d.display_key,
			d.sort_order, d.lastupdated, d.name_localization_id,
			c.id AS category_id, c.description AS category_description,
			COALESCE(c.local_abbrev, '') AS category_local_abbrev,
			c.name AS category_name, c.lastupdated AS category_lastupdated,
			COALESCE(NULLIF(lv.value, ''), d.dict_entry) AS sort_value`).
		Joins("JOIN clinlims.dictionary_category AS c ON c.id = d.dictionary_category_id").
		Joins("LEFT JOIN clinlims.localization_value AS lv ON lv.localization_id = d.name_localization_id AND lv.locale = 'en'").
		Where("c.name = ? AND d.is_active = ?", categoryName, "Y").
		// Byte order, matching Java's String.compareTo — see the note on
		// DictionaryByCategoryLocalizedSort.
		Order(`COALESCE(NULLIF(lv.value, ''), d.dict_entry) COLLATE "C"`).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// LocalizationRow is one clinlims.localization row.
type LocalizationRow struct {
	ID          string     `gorm:"column:id"`
	Description string     `gorm:"column:description"`
	Lastupdated *time.Time `gorm:"column:lastupdated"`
}

// LocalizationValueRow is one clinlims.localization_value row.
type LocalizationValueRow struct {
	ID             string     `gorm:"column:id"`
	LocalizationID string     `gorm:"column:localization_id"`
	Locale         string     `gorm:"column:locale"`
	Value          string     `gorm:"column:value"`
	LastUpdated    *time.Time `gorm:"column:last_updated"`
}

// LocalizationsByIDs fetches the localization rows and their per-locale values
// in two queries rather than one per dictionary entry.
func (d *DisplayListDAOImpl) LocalizationsByIDs(ids []string) (map[string]LocalizationRow, map[string][]LocalizationValueRow, error) {
	locs := map[string]LocalizationRow{}
	vals := map[string][]LocalizationValueRow{}
	if len(ids) == 0 {
		return locs, vals, nil
	}

	locRows := []LocalizationRow{}
	if err := d.DB.Table("clinlims.localization").
		Select("id, COALESCE(description, '') AS description, lastupdated").
		Where("id IN ?", ids).
		Scan(&locRows).Error; err != nil {
		return nil, nil, err
	}
	for _, r := range locRows {
		locs[r.ID] = r
	}

	valRows := []LocalizationValueRow{}
	if err := d.DB.Table("clinlims.localization_value").
		Select("id, localization_id, locale, value, last_updated").
		Where("localization_id IN ?", ids).
		Order("locale").
		Scan(&valRows).Error; err != nil {
		return nil, nil, err
	}
	for _, r := range valRows {
		vals[r.LocalizationID] = append(vals[r.LocalizationID], r)
	}
	return locs, vals, nil
}

// SupportedLocaleRow is one clinlims.supported_locale row.
//
// DisplayName is what feeds `localesAndValuesOfLocalesWithValues`
// ("English: Artibonite"), so it is the human label, not the code.
type SupportedLocaleRow struct {
	LocaleCode  string `gorm:"column:locale_code"`
	DisplayName string `gorm:"column:display_name"`
}

// ActiveSupportedLocales returns the active locales in display order — the
// source of allActiveLocales and localesSortedForDisplay.
func (d *DisplayListDAOImpl) ActiveSupportedLocales() ([]SupportedLocaleRow, error) {
	rows := []SupportedLocaleRow{}
	err := d.DB.Table("clinlims.supported_locale").
		Select("locale_code, display_name").
		Where("is_active = true").
		Order("sort_order").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}
