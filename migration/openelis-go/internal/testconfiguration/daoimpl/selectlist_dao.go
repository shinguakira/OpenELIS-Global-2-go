package daoimpl

import (
	"gorm.io/gorm"

	"openelis-go/internal/common/audittrail"
	locform "openelis-go/internal/localization/form"
)

// SelectListDAO backs TestRenameEntry, SelectListRenameEntry,
// ResultSelectListAdd and SaveResultSelectList.
type SelectListDAO struct {
	DB           *gorm.DB
	Audit        *audittrail.Service
	ActiveLocale string
}

func (d *SelectListDAO) locale() string {
	if d.ActiveLocale != "" {
		return d.ActiveLocale
	}
	return "en"
}

// RenameTest ports localizationService.updateTestNames: a test has TWO
// localizations — its name and its reporting name — and this screen writes both.
//
// Neither the test row nor its columns are touched, the same way the other
// rename screens leave their entity alone.
func (d *SelectListDAO) RenameTest(testID, nameEN, nameFR, reportEN, reportFR string) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		var ids []struct {
			Name      string `gorm:"column:name_id"`
			Reporting string `gorm:"column:reporting_id"`
		}
		if err := tx.Raw(`
			SELECT COALESCE(name_localization_id::text, '') AS name_id,
			       COALESCE(reporting_name_localization_id::text, '') AS reporting_id
			  FROM clinlims.test WHERE id = ?`, testID).Scan(&ids).Error; err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}
		ts := now()
		for _, pair := range []struct {
			id, en, fr string
		}{
			{ids[0].Name, nameEN, nameFR},
			{ids[0].Reporting, reportEN, reportFR},
		} {
			if pair.id == "" {
				continue
			}
			for locale, value := range map[string]string{"en": pair.en, "fr": pair.fr} {
				if err := tx.Exec(`
					UPDATE clinlims.localization_value
					   SET value = ?, last_updated = ?
					 WHERE localization_id = ? AND locale = ?`,
					value, ts, pair.id, locale).Error; err != nil {
					return err
				}
			}
			if err := tx.Exec(
				`UPDATE clinlims.localization SET lastupdated = ? WHERE id = ?`, ts, pair.id).Error; err != nil {
				return err
			}
		}

		// AND the test.name column, which localizationService.updateTestNames does
		// not appear to touch — it calls update() on the two Localizations and
		// nothing else. MEASURED anyway: renaming a test through this screen leaves
		// test.name holding the submitted English name. The Test entity is loaded
		// in the same transaction and flushed dirty; whatever syncs it, the column
		// moves, and the other rename screens do NOT do this.
		return tx.Exec(
			`UPDATE clinlims.test SET name = ?, lastupdated = ? WHERE id = ?`,
			nameEN, ts, testID).Error
	})
}

// SelectOption is one entry of resultSelectOptionList — the Dictionary ENTITY,
// as Jackson serialises it, with its Localization nested.
type SelectOption struct {
	Lastupdated             *int64                   `json:"lastupdated,omitempty"`
	NameKey                 *string                  `json:"nameKey,omitempty"`
	ID                      string                   `json:"id"`
	IsActive                string                   `json:"isActive"`
	DictEntry               string                   `json:"dictEntry"`
	DictionaryCategory      *DictionaryCategoryDTO   `json:"dictionaryCategory,omitempty"`
	LocalAbbreviation       *string                  `json:"localAbbreviation,omitempty"`
	SortOrder               *int                     `json:"sortOrder,omitempty"`
	LocalizedDictionaryName *locform.LocalizationDTO `json:"localizedDictionaryName,omitempty"`
	DisplayValue            string                   `json:"displayValue"`
}

// DictionaryCategoryDTO is the nested category on a Dictionary.
type DictionaryCategoryDTO struct {
	Lastupdated  *int64  `json:"lastupdated,omitempty"`
	ID           string  `json:"id"`
	CategoryName string  `json:"categoryName"`
	Description  *string `json:"description,omitempty"`
	LocalAbbrev  *string `json:"localAbbreviation,omitempty"`
}

// SelectOptionRow is the scan target for the select-option read.
type SelectOptionRow struct {
	ID                string  `gorm:"column:id"`
	IsActive          string  `gorm:"column:is_active"`
	DictEntry         string  `gorm:"column:dict_entry"`
	LocalAbbreviation *string `gorm:"column:local_abbrev"`
	SortOrder         *int    `gorm:"column:sort_order"`
	NameKey           *string `gorm:"column:name_key"`
	Lastupdated       *int64  `gorm:"column:lastupdated"`
	CategoryID        *string `gorm:"column:category_id"`
	CategoryName      *string `gorm:"column:category_name"`
	CategoryDesc      *string `gorm:"column:category_desc"`
	CategoryAbbrev    *string `gorm:"column:category_abbrev"`
	CategoryUpdated   *int64  `gorm:"column:category_updated"`
	LocalizationID    *string `gorm:"column:localization_id"`
	LocDescription    *string `gorm:"column:loc_description"`
	LocUpdated        *int64  `gorm:"column:loc_updated"`
}

// SelectListOptions ports getAllSelectListOptions.
//
// Not every dictionary entry: only the ones a DICTIONARY-VARIANT test result
// points at, de-duplicated, in the order getAllSortedTestResults returns them.
// So the list is driven by what tests actually offer as choices, not by the
// dictionary's own contents.
func (d *SelectListDAO) SelectListOptions() ([]SelectOptionRow, error) {
	rows := []SelectOptionRow{}
	err := d.DB.Raw(`
		SELECT DISTINCT ON (dict.id)
		       dict.id::text AS id,
		       COALESCE(dict.is_active, '') AS is_active,
		       COALESCE(dict.dict_entry, '') AS dict_entry,
		       dict.local_abbrev AS local_abbrev,
		       dict.sort_order::int AS sort_order,
		       dict.display_key AS name_key,
		       trunc(EXTRACT(EPOCH FROM dict.lastupdated) * 1000)::bigint AS lastupdated,
		       cat.id::text AS category_id,
		       cat.name AS category_name,
		       cat.description AS category_desc,
		       cat.local_abbrev AS category_abbrev,
		       trunc(EXTRACT(EPOCH FROM cat.lastupdated) * 1000)::bigint AS category_updated,
		       dict.name_localization_id::text AS localization_id,
		       loc.description AS loc_description,
		       trunc(EXTRACT(EPOCH FROM loc.lastupdated) * 1000)::bigint AS loc_updated
		  FROM clinlims.test_result tr
		  JOIN clinlims.dictionary dict ON dict.id::text = tr.value
		  LEFT JOIN clinlims.dictionary_category cat ON cat.id = dict.dictionary_category_id
		  LEFT JOIN clinlims.localization loc ON loc.id = dict.name_localization_id
		 WHERE tr.tst_rslt_type IN ('D', 'M')
		 ORDER BY dict.id`).Scan(&rows).Error
	return rows, err
}

// RenameSelectOption ports resultSelectListService.renameOption.
//
// It writes THREE things from one submitted English name: the localization's
// English value, dictionary.dict_entry and dictionary.local_abbrev. The French
// name reaches only the localization.
func (d *SelectListDAO) RenameSelectOption(optionID, nameEN, nameFR string, sysUserID int64) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		var locIDs []string
		if err := tx.Raw(
			`SELECT COALESCE(name_localization_id::text, '') FROM clinlims.dictionary WHERE id = ?`,
			optionID).Scan(&locIDs).Error; err != nil {
			return err
		}
		if len(locIDs) == 0 {
			return nil
		}
		ts := now()
		if locIDs[0] != "" {
			for locale, value := range map[string]string{"en": nameEN, "fr": nameFR} {
				if err := tx.Exec(`
					UPDATE clinlims.localization_value
					   SET value = ?, last_updated = ?
					 WHERE localization_id = ? AND locale = ?`,
					value, ts, locIDs[0], locale).Error; err != nil {
					return err
				}
			}
			if err := tx.Exec(
				`UPDATE clinlims.localization SET lastupdated = ? WHERE id = ?`, ts, locIDs[0]).Error; err != nil {
				return err
			}
		}
		var old []struct {
			DictEntry  string `gorm:"column:dict_entry"`
			LocalAbbev string `gorm:"column:local_abbrev"`
		}
		if err := tx.Raw(`SELECT COALESCE(dict_entry, '') AS dict_entry,
		                         COALESCE(local_abbrev, '') AS local_abbrev
		                    FROM clinlims.dictionary WHERE id = ?`, optionID).Scan(&old).Error; err != nil {
			return err
		}
		if len(old) == 0 || (old[0].DictEntry == nameEN && old[0].LocalAbbev == nameEN) {
			return nil
		}
		if err := tx.Exec(
			`UPDATE clinlims.dictionary SET dict_entry = ?, local_abbrev = ?, lastupdated = ? WHERE id = ?`,
			nameEN, nameEN, ts, optionID).Error; err != nil {
			return err
		}
		changes := ""
		if old[0].DictEntry != nameEN {
			changes += audittrail.Field("dictEntry", old[0].DictEntry)
		}
		if old[0].LocalAbbev != nameEN {
			changes += audittrail.Field("localAbbreviation", old[0].LocalAbbev)
		}
		return d.Audit.Write(tx, "DICTIONARY", optionID, sysUserID,
			audittrail.ActivityUpdate, &changes, ts)
	})
}

// TestResultItem is one entry of the testSelectListJson `items` array.
type TestResultItem struct {
	// ID names an EXISTING test_result to re-order; its absence means a new one.
	ID          *string
	Order       int
	Normal      bool
	Qualifiable bool
}

// AddSelectList ports addResultSelectList: a new dictionary entry plus a
// test_result row per test that should offer it.
//
// The sort order is stored as the submitted one TIMES TEN, on both branches —
// the same multiplier TestActivation applies.
func (d *SelectListDAO) AddSelectList(nameEN, nameFR, loinc string, byTest map[string][]TestResultItem, sysUserID int64) (string, error) {
	var dictID string
	err := d.DB.Transaction(func(tx *gorm.DB) error {
		ts := now()

		var locID string
		if err := tx.Raw(`
			INSERT INTO clinlims.localization (id, description, lastupdated)
			VALUES (nextval('clinlims.localization_seq'), NULL, ?)
			RETURNING id::text`, ts).Scan(&locID).Error; err != nil {
			return err
		}
		for locale, value := range map[string]string{"en": nameEN, "fr": nameFR} {
			if err := tx.Exec(`
				INSERT INTO clinlims.localization_value (id, localization_id, locale, value, last_updated)
				VALUES (nextval('clinlims.localization_value_seq'), ?, ?, ?, ?)`,
				locID, locale, value, ts).Error; err != nil {
				return err
			}
		}

		// sortOrder 1, is_active 'Y', dict_entry and local_abbrev both the
		// English name, category "Test Result".
		if err := tx.Raw(`
			INSERT INTO clinlims.dictionary
			       (id, dict_entry, local_abbrev, is_active, sort_order, dictionary_category_id,
			        name_localization_id, lastupdated)
			VALUES (nextval('clinlims.dictionary_seq'), ?, ?, 'Y', 1,
			        (SELECT id FROM clinlims.dictionary_category WHERE name = 'Test Result' LIMIT 1),
			        ?, ?)
			RETURNING id::text`, nameEN, nameEN, locID, ts).Scan(&dictID).Error; err != nil {
			return err
		}

		for testID, items := range byTest {
			for _, item := range items {
				if item.ID != nil {
					// An existing test_result is only RE-ORDERED; its value and
					// flags are left alone.
					if err := tx.Exec(`
						UPDATE clinlims.test_result SET sort_order = ?, lastupdated = ?
						 WHERE test_id = ? AND value = ?`,
						10*item.Order, ts, testID, *item.ID).Error; err != nil {
						return err
					}
					continue
				}
				if err := tx.Exec(`
					INSERT INTO clinlims.test_result
					       (id, test_id, tst_rslt_type, value, sort_order, is_quantifiable,
					        is_normal, result_group, lastupdated)
					-- result_group is a NUMERIC column and setResultGroup("") maps to
					-- NULL, not to a zero: an empty string is not a number.
					VALUES (nextval('clinlims.test_result_seq'), ?, 'D', ?, ?, ?, ?, NULL, ?)`,
					testID, dictID, item.Order*10,
					item.Qualifiable, item.Normal, ts).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	return dictID, err
}

// TestDictionaryRow is one dictionary option a test already offers.
type TestDictionaryRow struct {
	TestID       string `gorm:"column:test_id"`
	DictionaryID string `gorm:"column:dictionary_id"`
	Value        string `gorm:"column:value"`
}

// TestSelectDictionary ports resultSelectListService.getTestSelectDictionary:
// per test, the dictionary entries its dictionary-variant results point at, in
// sort order.
func (d *SelectListDAO) TestSelectDictionary() ([]TestDictionaryRow, error) {
	rows := []TestDictionaryRow{}
	err := d.DB.Table("clinlims.test_result AS tr").
		Select(`tr.test_id::text AS test_id,
		        dict.id::text AS dictionary_id,
		        COALESCE(NULLIF(lv.value, ''), dict.dict_entry) AS value`).
		Joins("JOIN clinlims.dictionary AS dict ON dict.id::text = tr.value").
		Joins(`LEFT JOIN clinlims.localization_value AS lv
		         ON lv.localization_id = dict.name_localization_id AND lv.locale = ?`, d.locale()).
		Where("tr.tst_rslt_type IN ('D', 'M')").
		Order("tr.test_id, tr.sort_order NULLS LAST, tr.id").
		Scan(&rows).Error
	return rows, err
}
