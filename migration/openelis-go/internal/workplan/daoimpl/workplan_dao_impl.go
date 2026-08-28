// Package daoimpl ports the analysis queries behind rest/WorkPlanBy*
// (constitution.md Layer II). Folder layout mirrors the Java source.
package daoimpl

import (
	"strings"

	"gorm.io/gorm"
)

// WorkplanDAOImpl backs the four WorkPlanBy* routes (Wave 5.5-5.8).
type WorkplanDAOImpl struct {
	DB *gorm.DB

	// ActiveLocale is site_information "default language locale". Empty falls
	// back to "en".
	ActiveLocale string
}

// Locale returns the configured locale, or "en" when unset.
func (d *WorkplanDAOImpl) Locale() string {
	if d.ActiveLocale != "" {
		return d.ActiveLocale
	}
	return "en"
}

// workplanStatusNames are the four analysis statuses WorkplanRestController
// collects in its @PostConstruct, spelled as they are STORED in
// status_of_sample — StatusService.addToAnalysisMap matches the name
// literally, so AnalysisStatus.NotStarted is the row named "Not Tested" and
// BiologistRejected is "Biologist Rejection".
//
// Resolved by name rather than by id because the ids are deployment data.
var workplanStatusNames = []string{
	"Not Tested",
	"Biologist Rejection",
	"Technical Rejected",
	"NonConforming",
}

// WorkplanAnalysisRow is one analysis as the workplan builders read it.
type WorkplanAnalysisRow struct {
	AnalysisID      int64   `gorm:"column:analysis_id"`
	TestID          string  `gorm:"column:test_id"`
	TestName        string  `gorm:"column:test_name"`
	AccessionNumber string  `gorm:"column:accession_number"`
	ReceivedDate    *string `gorm:"column:received_date"`
}

// selectClause is shared by all four queries.
//
// received_date is formatted IN SQL as dd/MM/yyyy HH24:MI, which is what
// getReceivedDateDisplay produces. Doing it here rather than in Go keeps the
// value in the database's own timezone, the same one Java's Timestamp renders
// in — formatting a driver-parsed time in the Go process picks up whatever
// zone that process runs in instead.
const selectClause = `a.id AS analysis_id,
	a.test_id::text AS test_id,
	COALESCE(lv.value, t.name) AS test_name,
	s.accession_number AS accession_number,
	to_char(s.received_date, 'DD/MM/YYYY HH24:MI') AS received_date`

func (d *WorkplanDAOImpl) base() *gorm.DB {
	return d.DB.Table("clinlims.analysis AS a").
		Select(selectClause).
		Joins("JOIN clinlims.sample_item AS si ON si.id = a.sampitem_id").
		Joins("JOIN clinlims.sample AS s ON s.id = si.samp_id").
		Joins("JOIN clinlims.test AS t ON t.id = a.test_id").
		Joins("LEFT JOIN clinlims.localization AS l ON l.id = t.name_localization_id").
		Joins("LEFT JOIN clinlims.localization_value AS lv ON lv.localization_id = l.id AND lv.locale = ?", d.Locale()).
		Joins(`JOIN clinlims.status_of_sample AS st ON st.id = a.status_id
			AND st.status_type = 'ANALYSIS' AND st.name IN ?`, workplanStatusNames)
}

// ByTest ports getAllAnalysisByTestAndStatus:
//
//	from Analysis a where a.test.id = :testId and a.statusId IN (:statusIdList)
//	order by a.sampleItem.sample.accessionNumber
//
// The ORDER BY is on the accession, which sorts under the DATABASE collation —
// on this deployment that collation ignores punctuation, so E2E001 comes
// FIRST. Java then re-sorts the list in memory with String.compareTo, which is
// byte order and puts E2E001 LAST. Both orderings are load-bearing: this one
// decides the sampleGroupingNumber each row is stamped with, the in-memory one
// decides the array order. Sorting here the way Java sorts later would assign
// the wrong grouping numbers.
func (d *WorkplanDAOImpl) ByTest(testID string) ([]WorkplanAnalysisRow, error) {
	rows := []WorkplanAnalysisRow{}
	err := d.base().
		Where("a.test_id = ?", testID).
		Order("s.accession_number").
		Scan(&rows).Error
	return rows, err
}

// ByTestSection ports getAllAnalysisByTestSectionAndStatus:
//
//	from Analysis a where a.testSection.id = :testSectionId and a.statusId IN (...)
//	order by a.id
//
// a.testSection is analysis.test_sect_id — the DENORMALISED column, not the
// section reached through test. The two normally agree because
// AnalysisServiceImpl copies one into the other at creation time, so a port
// that joined through test would look correct until they diverge.
//
// The method's third parameter, sortedByDateAndAccession, guards a block whose
// entire body is commented out. Callers pass true and it does nothing.
func (d *WorkplanDAOImpl) ByTestSection(sectionID string) ([]WorkplanAnalysisRow, error) {
	rows := []WorkplanAnalysisRow{}
	err := d.base().
		Where("a.test_sect_id = ?", sectionID).
		Order("a.id").
		Scan(&rows).Error
	return rows, err
}

// PanelMemberTests ports panelItemService.getPanelItemsForPanel.
//
// WorkPlanByPanel does NOT filter analyses on analysis.panel_id: it reads the
// panel's member tests and runs the ByTest query once per member, concatenating
// the results. An analysis on a member test therefore appears even with its own
// panel_id NULL, and the response is strictly larger than the set of analyses
// carrying that panel_id.
func (d *WorkplanDAOImpl) PanelMemberTests(panelID string) ([]string, error) {
	ids := []string{}
	err := d.DB.Table("clinlims.panel_item").
		Select("test_id::text").
		Where("panel_id = ?", panelID).
		Order("ctid").
		Scan(&ids).Error
	return ids, err
}

// ByPriority ports getAnalysesByPriorityAndStatusId:
//
//	from Analysis a where a.sampleItem.sample.priority = :oderpriority
//	and a.statusId in (:statusIds)
//
// No ORDER BY at all, so the order is whatever the plan yields — and it is
// observable, because the grouping counter is stamped in THIS order before the
// list is sorted.
//
// Measured against live Java: the order is SAMPLE_ITEM physical order, not
// analysis order. Hibernate drives the join from sample_item, so ORDER BY
// si.ctid reproduces it; a.ctid does not (it puts E2E001 first where Java puts
// it sixth, shifting every grouping number by one).
func (d *WorkplanDAOImpl) ByPriority(priority string) ([]WorkplanAnalysisRow, error) {
	rows := []WorkplanAnalysisRow{}
	err := d.base().
		Where("s.order_priority = ?", priority).
		Order("si.ctid").
		Scan(&rows).Error
	return rows, err
}

// OrderPriorities are the OrderPriority enum constants. Spring binds the
// `priority` request parameter to that enum, so a value outside this set is a
// BINDING failure (400) rather than an empty result — the only one of the four
// WorkPlan endpoints that can 400 on a bad parameter.
func OrderPriorities() []string {
	return []string{"ROUTINE", "ASAP", "STAT", "TIMED", "FUTURE_STAT"}
}

// IsOrderPriority reports whether v is a member of the enum.
func IsOrderPriority(v string) bool {
	for _, p := range OrderPriorities() {
		if strings.EqualFold(p, v) {
			return p == v
		}
	}
	return false
}

// AugmentedTestName ports TestServiceImpl.buildAugmentedTestNameForLocale:
// the localized test name, plus "(<localized type of sample>)" when
// TEST_NAME_AUGMENTED is true and the test's FIRST type_of_sample_test row is
// not the "Variable" pseudo-type.
//
// Only WorkPlanByTestSection uses this form — it goes through
// AnalysisServiceImpl.getTestDisplayName. ByPanel and ByPriority call
// getUserLocalizedTestName, which is buildTestName and returns the BARE
// localized name. Same conceptual field, two builders, and they differ
// whenever a test has a sample type.
func (d *WorkplanDAOImpl) augmentedNameSelect() string {
	return `COALESCE(lv.value, t.name) || COALESCE(
		(SELECT '(' || COALESCE(tlv.value, tos.description) || ')'
		   FROM clinlims.sampletype_test AS tost
		   JOIN clinlims.type_of_sample AS tos ON tos.id = tost.sample_type_id
		   LEFT JOIN clinlims.localization AS tl ON tl.id = tos.name_localization_id
		   LEFT JOIN clinlims.localization_value AS tlv
		          ON tlv.localization_id = tl.id AND tlv.locale = '` + d.Locale() + `'
		  WHERE tost.test_id = t.id
		    AND tos.local_abbrev IS DISTINCT FROM 'Variable'
		  ORDER BY tost.ctid LIMIT 1), '')`
}

// SectionHasTypelessItem reports whether the section has an analysis sitting on
// a sample_item with a NULL typeosamp_id.
//
// AnalysisServiceImpl.getTestDisplayName dereferences
// sampleItem.getTypeOfSampleId() with no null check, so ONE such analysis makes
// the whole request 500 — after the query has already succeeded. This is a JAVA
// DEFECT reproduced deliberately: the port must fail on exactly the same
// inputs, so the condition is checked rather than the null tolerated.
//
// Java's own unassigned-sample HQL LEFT JOINs type_of_sample and COALESCEs the
// description, so the two code paths disagree about whether the state is legal.
func (d *WorkplanDAOImpl) SectionHasTypelessItem(sectionID string) (bool, error) {
	var n int64
	err := d.DB.Table("clinlims.analysis AS a").
		Joins("JOIN clinlims.sample_item AS si ON si.id = a.sampitem_id").
		Joins(`JOIN clinlims.status_of_sample AS st ON st.id = a.status_id
			AND st.status_type = 'ANALYSIS' AND st.name IN ?`, workplanStatusNames).
		Where("a.test_sect_id = ? AND si.typeosamp_id IS NULL", sectionID).
		Count(&n).Error
	return n > 0, err
}

// ByTestSectionAugmented is ByTestSection with the augmented test name, which
// is the form that endpoint emits.
func (d *WorkplanDAOImpl) ByTestSectionAugmented(sectionID string) ([]WorkplanAnalysisRow, error) {
	rows := []WorkplanAnalysisRow{}
	err := d.DB.Table("clinlims.analysis AS a").
		Select(`a.id AS analysis_id,
			a.test_id::text AS test_id,
			`+d.augmentedNameSelect()+` AS test_name,
			s.accession_number AS accession_number,
			to_char(s.received_date, 'DD/MM/YYYY HH24:MI') AS received_date`).
		Joins("JOIN clinlims.sample_item AS si ON si.id = a.sampitem_id").
		Joins("JOIN clinlims.sample AS s ON s.id = si.samp_id").
		Joins("JOIN clinlims.test AS t ON t.id = a.test_id").
		Joins("LEFT JOIN clinlims.localization AS l ON l.id = t.name_localization_id").
		Joins("LEFT JOIN clinlims.localization_value AS lv ON lv.localization_id = l.id AND lv.locale = ?", d.Locale()).
		Joins(`JOIN clinlims.status_of_sample AS st ON st.id = a.status_id
			AND st.status_type = 'ANALYSIS' AND st.name IN ?`, workplanStatusNames).
		Where("a.test_sect_id = ?", sectionID).
		Order("a.id").
		Scan(&rows).Error
	return rows, err
}
