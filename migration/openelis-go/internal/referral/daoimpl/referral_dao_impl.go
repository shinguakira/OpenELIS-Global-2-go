// Package daoimpl ports the referral queries behind rest/ReferredOutTests
// (constitution.md Layer II). Folder layout mirrors the Java source.
package daoimpl

import (
	"fmt"
	"strings"

	"openelis-go/internal/common/i18n"
	"openelis-go/internal/common/util"

	"gorm.io/gorm"
)

// noteTypes are Note.INTERNAL/EXTERNAL/REJECT_REASON/NON_CONFORMITY, in the
// order getNotePrefix tests them. Fixed order because the slice builds SQL and
// a map would emit a different string on every run.
var noteTypes = []struct{ Code, Key string }{
	{"I", "note.type.internal"},
	{"E", "note.type.external"},
	{"R", "note.type.rejectReason"},
	{"N", "note.type.nonConformity"},
}

// notePrefixCase is getNotePrefix as a SQL expression: the localized label for
// a known note type, the empty string for anything else.
//
// The labels come from the message bundle rather than being spelled out here,
// because they ARE message keys in Java and a port that hardcodes "Internal"
// has quietly made the field untranslatable.
func notePrefixCase() string {
	msgs := i18n.Messages()
	var b strings.Builder
	b.WriteString("CASE n.note_type")
	for _, t := range noteTypes {
		fmt.Fprintf(&b, " WHEN '%s' THEN '%s'", t.Code, strings.ReplaceAll(msgs[t.Key], "'", "''"))
	}
	b.WriteString(" ELSE '' END")
	return b.String()
}

// ReferralDAOImpl backs rest/ReferredOutTests (Wave 5.4).
type ReferralDAOImpl struct {
	DB *gorm.DB

	// ActiveLocale is site_information "default language locale".
	ActiveLocale string
}

// Locale returns the configured locale, or "en" when unset.
func (d *ReferralDAOImpl) Locale() string {
	if d.ActiveLocale != "" {
		return d.ActiveLocale
	}
	return "en"
}

// ReferralDisplayRow carries everything convertToDisplayItem reads.
type ReferralDisplayRow struct {
	AccessionNumber  string  `gorm:"column:accession_number"`
	ReferredSendDate *string `gorm:"column:referred_send_date"`
	Status           *string `gorm:"column:status"`
	PatientLastName  *string `gorm:"column:patient_last_name"`
	PatientFirstName *string `gorm:"column:patient_first_name"`
	TestName         *string `gorm:"column:test_name"`
	OrganizationName *string `gorm:"column:organization_name"`
	AnalysisID       string  `gorm:"column:analysis_id"`
	// notes is getNotesAsString(analysis, prefixType=true, prefixTimestamp=true,
	// "<br/>", excludeExternPrefix=false), which is NOT the bare note text:
	//
	//     <type label> <dd/MM/yyyy HH:mm> : <text>
	//
	// notesToString appends the type, a space, the timestamp, a space, then the
	// literal ": " — so an unknown note type, whose label is the empty string,
	// still contributes its space and the line opens with one.
	//
	// Chronological by lastupdated, per getNotesChronologicallyByRefIdAndRefTable.
	//
	// The referral's returned result and the analysis notes. convertToDisplayItem
	// sets referralResultsDisplay and resultDate only inside
	// `if (!resultList.isEmpty())`, and notes come from getNotesAsString — every
	// seeded referral had neither, so all three fields were unreachable.
	ResultValue   *string `gorm:"column:result_value"`
	ResultType    *string `gorm:"column:result_type"`
	UnitOfMeasure *string `gorm:"column:unit_of_measure"`
	CompletedDate *string `gorm:"column:completed_date"`
	Notes         *string `gorm:"column:notes"`
}

// ByAccessionNumber ports getReferralsByAccessionNumber, then the per-referral
// work convertToDisplayItem does, as one query.
//
// referringTestName is the BARE localized test name
// (test.getLocalizedTestName().getLocalizedValue()) — NOT the augmented form
// with the sample type that the same wave's WorkPlanByTestSection emits and
// that this endpoint's own testSelectionList uses. Three name builders in one
// wave; they are not interchangeable.
func (d *ReferralDAOImpl) ByAccessionNumber(accessionNumber string) ([]ReferralDisplayRow, error) {
	rows := []ReferralDisplayRow{}
	err := d.DB.Table("clinlims.referral AS r").
		Select(`s.accession_number AS accession_number,
			to_char(r.sent_date, 'DD/MM/YYYY') AS referred_send_date,
			r.status AS status,
			pe.last_name  AS patient_last_name,
			pe.first_name AS patient_first_name,
			COALESCE(lv.value, t.name) AS test_name,
			o.name AS organization_name,
			a.id::text AS analysis_id,
			res.value AS result_value,
			res.result_type AS result_type,
			uom.name AS unit_of_measure,
			to_char(a.completed_date, 'DD/MM/YYYY') AS completed_date,
			(SELECT string_agg(`+notePrefixCase()+` || ' '
			          || to_char(n.lastupdated, 'DD/MM/YYYY HH24:MI') || ' : ' || n.text,
			        '<br/>' ORDER BY n.lastupdated)
			   FROM clinlims.note n
			   JOIN clinlims.reference_tables rt ON rt.id = n.reference_table
			  WHERE rt.name = 'ANALYSIS' AND n.reference_id = a.id) AS notes`).
		Joins("JOIN clinlims.analysis AS a ON a.id = r.analysis_id").
		Joins("JOIN clinlims.sample_item AS si ON si.id = a.sampitem_id").
		Joins("JOIN clinlims.sample AS s ON s.id = si.samp_id").
		Joins("JOIN clinlims.test AS t ON t.id = a.test_id").
		Joins("LEFT JOIN clinlims.localization AS l ON l.id = t.name_localization_id").
		Joins("LEFT JOIN clinlims.localization_value AS lv ON lv.localization_id = l.id AND lv.locale = ?", d.Locale()).
		Joins("LEFT JOIN clinlims.organization AS o ON o.id = r.organization_id").
		Joins("LEFT JOIN clinlims.sample_human AS sh ON sh.samp_id = s.id").
		Joins("LEFT JOIN clinlims.patient AS pa ON pa.id = sh.patient_id").
		Joins("LEFT JOIN clinlims.person AS pe ON pe.id = pa.person_id").
		// res and uom come LAST: they reference a and t, and Postgres needs the
		// aliases defined earlier in the FROM chain.
		Joins("LEFT JOIN clinlims.result AS res ON res.analysis_id = a.id").
		Joins("LEFT JOIN clinlims.unit_of_measure AS uom ON uom.id = t.uom_id").
		Where("s.accession_number = ?", accessionNumber).
		Order("r.ctid").
		Scan(&rows).Error
	return rows, err
}

// AllActiveTests ports DisplayListService.createTestList — every ACTIVE test,
// valued by the AUGMENTED localized name (name plus "(sample type)"), sorted by
// that value with Java's String.compareTo.
//
// The sort is done in SQL with COLLATE "C", which is byte order and therefore
// the same comparison Java performs; the database's own collation ignores
// punctuation and would order this list differently.
func (d *ReferralDAOImpl) AllActiveTests() ([]util.IdValuePair, error) {
	rows := []util.IdValuePair{}
	// The locale is BOUND, not concatenated. It comes from site_information
	// rather than from a caller, so this is not a live injection path — but the
	// value still reaches the query as SQL text, and the sibling result DAO
	// already shows the parameterised form. Keeping one pattern is cheaper than
	// arguing about which strings are safe.
	augmented := `COALESCE(lv.value, t.name) || COALESCE(
		(SELECT '(' || COALESCE(tlv.value, tos.description) || ')'
		   FROM clinlims.sampletype_test AS tost
		   JOIN clinlims.type_of_sample AS tos ON tos.id = tost.sample_type_id
		   LEFT JOIN clinlims.localization AS tl ON tl.id = tos.name_localization_id
		   LEFT JOIN clinlims.localization_value AS tlv
		          ON tlv.localization_id = tl.id AND tlv.locale = @loc
		  WHERE tost.test_id = t.id
		    AND tos.local_abbrev IS DISTINCT FROM 'Variable'
		  ORDER BY tost.ctid LIMIT 1), '')`
	// The expression is needed twice — once to produce the value and once to
	// sort it — but a named parameter can only be bound on the SELECT. Ordering
	// by the output alias directly is not an option either: Postgres accepts a
	// bare output name in ORDER BY but not one inside an expression, so
	// `ORDER BY value COLLATE "C"` is a missing-column error.
	//
	// A derived table solves both: the locale binds once, and the outer query
	// sorts a real column.
	inner := d.DB.Table("clinlims.test AS t").
		Select("t.id::text AS id, "+augmented+" AS value", map[string]any{"loc": d.Locale()}).
		Joins("LEFT JOIN clinlims.localization AS l ON l.id = t.name_localization_id").
		Joins("LEFT JOIN clinlims.localization_value AS lv ON lv.localization_id = l.id AND lv.locale = ?", d.Locale()).
		Where("t.is_active = ?", "Y")
	err := d.DB.Table("(?) AS x", inner).
		Select("x.id, x.value").
		Order(`x.value COLLATE "C"`).
		Scan(&rows).Error
	return rows, err
}

// TestSectionsByName ports DisplayListService.createTestSectionByNameList.
//
// This is NOT the same list as ListType.TEST_SECTION, even though both are
// "the active test sections":
//
//	TEST_SECTION          COALESCE(localization value, name)  -- localized
//	TEST_SECTION_BY_NAME  section.getTestSectionName()        -- the RAW column
//
// The two disagree wherever a section's stored name is in a different language
// from its localization. On this deployment section 76 is stored as "Virologie"
// with an English localization of "Virology", so reusing the localized list
// here emits "Virology" where Java emits "Virologie" — one row out of ten, and
// invisible on the other nine.
func (d *ReferralDAOImpl) TestSectionsByName() ([]util.IdValuePair, error) {
	rows := []util.IdValuePair{}
	err := d.DB.Table("clinlims.test_section AS ts").
		Select("ts.id::text AS id, ts.name AS value").
		Where("ts.is_active = ?", "Y").
		Order("ts.sort_order").
		Scan(&rows).Error
	return rows, err
}
