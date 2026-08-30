package daoimpl

import (
	"sort"
	"strings"

	"openelis-go/internal/common/util"
)

// The DisplayListService lists TestAdd's GET reads that no earlier wave needed,
// plus the two rows-only reads its service assembles itself.

// UnitsOfMeasure ports createUOMList: UNIT_OF_MEASURE.
//
// unitOfMeasureService.getAll() — EVERY unit, active or not. UNIT_OF_MEASURE,
// UNIT_OF_MEASURE_ACTIVE and UNIT_OF_MEASURE_INACTIVE are all registered
// against this one builder, so the three list types hold identical rows and two
// of the names lie.
//
// The displayed value is getLocalizedName(), and unit_of_measure has no
// name_key column — so nameKey is null and it falls through to
// getDefaultLocalizedName(), which UnitOfMeasure overrides to return the plain
// `name`. No localization join: this list is the same in every locale.
func (d *DisplayListDAOImpl) UnitsOfMeasure() ([]util.IdValuePair, error) {
	rows := []idValueRow{}
	err := d.DB.Table("clinlims.unit_of_measure AS u").
		Select("u.id AS id, COALESCE(u.name, '') AS value").
		Order("u.id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return toPairs(rows), nil
}

// ResultTypeRow is one type_of_test_result, by DESCRIPTION — which is what
// createLocalizedResultTypeList branches on.
type ResultTypeRow struct {
	ID          string `gorm:"column:id"`
	Description string `gorm:"column:description"`
}

// ResultTypes reads type_of_test_result for RESULT_TYPE_LOCALIZED.
func (d *DisplayListDAOImpl) ResultTypes() ([]ResultTypeRow, error) {
	rows := []ResultTypeRow{}
	err := d.DB.Table("clinlims.type_of_test_result AS t").
		Select("t.id::text AS id, COALESCE(t.description, '') AS description").
		Order("t.id").
		Scan(&rows).Error
	return rows, err
}

// dictionaryTestResultCategories is createDictionaryTestResults' source list,
// in the order it concatenates them. The order does not survive — the whole
// result is re-sorted by lowercased value — but a category missing from it
// drops its entries entirely.
var dictionaryTestResultCategories = []string{
	"CG", "HL", "KL", "Test Result", "HIV1NInd", "PosNegIndInv", "HIVResult",
}

// DictionaryTestResults ports createDictionaryTestResults: DICTIONARY_TEST_RESULTS.
//
// Seven categories concatenated, then sorted by value.toLowerCase() with
// String.compareTo — case-folded first and BYTE order after, which is neither
// the database's default collation nor the `COLLATE "C"` the per-category query
// uses. An entry belonging to two of the categories appears twice; there is no
// dedup.
func (d *DisplayListDAOImpl) DictionaryTestResults() ([]util.IdValuePair, error) {
	out := []util.IdValuePair{}
	for _, category := range dictionaryTestResultCategories {
		pairs, err := d.DictionaryByCategoryLocalizedSort(category)
		if err != nil {
			return nil, err
		}
		out = append(out, pairs...)
	}
	// Collections.sort is stable, so equal lowercased values keep the
	// concatenation order — category by category, each already localized-sorted.
	sort.SliceStable(out, func(i, j int) bool {
		return strings.ToLower(out[i].Value) < strings.ToLower(out[j].Value)
	})
	return out, nil
}

// AgeRangeRow is one resultAgeRange site_information row.
//
// The pair the form carries is INVERTED against every other list here: the id
// is the site information VALUE (an age in months, or the literal "Infinity")
// and the displayed value is its localized NAME.
type AgeRangeRow struct {
	Value   string `gorm:"column:value"`
	Name    string `gorm:"column:name"`
	NameKey string `gorm:"column:name_key"`
}

// PredefinedAgeRanges reads ResultLimitServiceImpl.getPredefinedAgeRanges'
// source rows. The name resolution and the sort live in the service, because
// the first is a message lookup.
func (d *DisplayListDAOImpl) PredefinedAgeRanges() ([]AgeRangeRow, error) {
	rows := []AgeRangeRow{}
	err := d.DB.Table("clinlims.site_information AS si").
		Select(`COALESCE(si.value, '') AS value,
		        COALESCE(si.name, '') AS name,
		        COALESCE(si.name_key, '') AS name_key`).
		Joins("JOIN clinlims.site_information_domain AS dom ON dom.id = si.domain_id").
		Where("dom.name = ?", "resultAgeRange").
		Order("si.id").
		Scan(&rows).Error
	return rows, err
}

// SortAgeRanges is getPredefinedAgeRanges' comparator: "Infinity" last,
// everything else by the id read as an integer.
//
// Java's comparator calls Integer.parseInt on any id that is not "Infinity", so
// a resultAgeRange row whose value is neither numeric nor "Infinity" is a 500.
// Here it sorts as zero — the deployment ships five rows and none of them is
// malformed, so the difference is unreachable.
func SortAgeRanges(pairs []util.IdValuePair) {
	sort.SliceStable(pairs, func(i, j int) bool {
		if pairs[i].Id == "Infinity" {
			return false
		}
		if pairs[j].Id == "Infinity" {
			return true
		}
		return atoiOrZero(pairs[i].Id) < atoiOrZero(pairs[j].Id)
	})
}

// DictionaryResultGroupRow is one dictionary-variant test_result, in the order
// getAllSortedTestResults returns it.
type DictionaryResultGroupRow struct {
	TestID string `gorm:"column:test_id"`
	Value  string `gorm:"column:value"`
	Name   string `gorm:"column:name"`
	Found  bool   `gorm:"column:found"`
}

// DictionaryResultGroups ports createGroupedDictionaryList's source read.
//
// getAllSortedTestResults orders by the test id AS A STRING and then by sort
// order with nulls first, which is all the grouping needs: it cuts a new group
// whenever the test id changes, so only adjacency matters. Rows that are not a
// dictionary variant, or whose value is blank, take no part.
//
// The lookup is getDictionaryById, so a value naming no dictionary row is
// dropped from its group — the group survives one member shorter.
func (d *DisplayListDAOImpl) DictionaryResultGroups() ([]DictionaryResultGroupRow, error) {
	rows := []DictionaryResultGroupRow{}
	err := d.DB.Table("clinlims.test_result AS tr").
		Select(`tr.test_id::text AS test_id,
		        COALESCE(tr.value, '') AS value,
		        COALESCE(NULLIF(lv.value, ''), dict.dict_entry, '') AS name,
		        (dict.id IS NOT NULL) AS found`).
		Joins("LEFT JOIN clinlims.dictionary AS dict ON dict.id::text = tr.value").
		Joins(`LEFT JOIN clinlims.localization_value AS lv
		         ON lv.localization_id = dict.name_localization_id AND lv.locale = ?`, d.Locale()).
		Where("tr.tst_rslt_type IN ('D', 'M', 'C') AND COALESCE(btrim(tr.value), '') <> ''").
		Order("tr.test_id::text, tr.sort_order NULLS FIRST, tr.id").
		Scan(&rows).Error
	return rows, err
}

func atoiOrZero(s string) int {
	n, neg := 0, false
	for i, c := range s {
		if i == 0 && c == '-' {
			neg = true
			continue
		}
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	if neg {
		return -n
	}
	return n
}
