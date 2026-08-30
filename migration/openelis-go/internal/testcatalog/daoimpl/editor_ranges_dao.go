package daoimpl

import (
	"math"
	"strconv"
	"time"

	"gorm.io/gorm"

	"openelis-go/internal/common/audittrail"
	"openelis-go/internal/common/util"
)

// The Reference Ranges section — the only part of the editor that decides
// whether a result reads as normal, and the only part of it that IS audited.
//
// RESULT_LIMITS carries keep_history='Y' and ResultLimitServiceImpl honours it:
// an insert leaves an 'I' with a null payload, a delete leaves a 'D' carrying
// the whole row in the entity's declared-field order. An UPDATE leaves nothing —
// the same asymmetry TestModifyEntry measured.

// numericResultTypeID is ResultLimitServiceImpl.NUMERIC_RESULT_TYPE_ID,
// resolved from type_of_test_result at startup in Java. The Ranges editor
// manages ONLY rows carrying it; a dictionary limit is invisible to the diff and
// therefore survives a save that does not mention it.
const numericResultTypeID = "4"

// RangeRow is one reference range as the editor reads it.
type RangeRow struct {
	ID                 string  `gorm:"column:id"`
	ResultTypeID       string  `gorm:"column:result_type_id"`
	ComponentID        *string `gorm:"column:component_id"`
	Gender             *string `gorm:"column:gender"`
	MinAge             float64 `gorm:"column:min_age"`
	MaxAge             float64 `gorm:"column:max_age"`
	LowNormal          float64 `gorm:"column:low_normal"`
	HighNormal         float64 `gorm:"column:high_normal"`
	LowCritical        float64 `gorm:"column:low_critical"`
	HighCritical       float64 `gorm:"column:high_critical"`
	LowValid           float64 `gorm:"column:low_valid"`
	HighValid          float64 `gorm:"column:high_valid"`
	LowReporting       float64 `gorm:"column:low_reporting_range"`
	HighReporting      float64 `gorm:"column:high_reporting_range"`
	DictionaryNormalID *string `gorm:"column:normal_dictionary_id"`
	AlwaysValidate     bool    `gorm:"column:always_validate"`
	Lastupdated        *string `gorm:"column:lastupdated"`
}

// Ranges reads every result limit of a test — including the dictionary ones,
// because the GET lists them and the coverage report counts them.
func (d *EditorTestDAO) Ranges(testID string) ([]RangeRow, error) {
	rows := []RangeRow{}
	err := d.DB.Raw(rangeSelect+` WHERE test_id = ? ORDER BY id`, testID).Scan(&rows).Error
	return rows, err
}

const rangeSelect = `
	SELECT id::text AS id, test_result_type_id::text AS result_type_id,
	       component_id, gender,
	       COALESCE(min_age, 0) AS min_age,
	       COALESCE(max_age, 'Infinity'::float8) AS max_age,
	       COALESCE(low_normal, '-Infinity'::float8) AS low_normal,
	       COALESCE(high_normal, 'Infinity'::float8) AS high_normal,
	       COALESCE(low_critical, 'Infinity'::float8) AS low_critical,
	       COALESCE(high_critical, 'Infinity'::float8) AS high_critical,
	       COALESCE(low_valid, '-Infinity'::float8) AS low_valid,
	       COALESCE(high_valid, 'Infinity'::float8) AS high_valid,
	       COALESCE(low_reporting_range, '-Infinity'::float8) AS low_reporting_range,
	       COALESCE(high_reporting_range, 'Infinity'::float8) AS high_reporting_range,
	       normal_dictionary_id::text AS normal_dictionary_id,
	       COALESCE(always_validate, false) AS always_validate,
	       to_char(lastupdated, 'YYYY-MM-DD HH24:MI:SS.MS') AS lastupdated
	  FROM clinlims.result_limits`

// DesiredRange is one range the caller asked for. The bounds are already
// unboxed: a null in the request became the entity default before it got here.
type DesiredRange struct {
	ID           string
	ComponentID  *string
	Gender       *string
	MinAge       float64
	MaxAge       float64
	LowNormal    float64
	HighNormal   float64
	LowCritical  float64
	HighCritical float64
	LowValid     float64
	HighValid    float64
}

// SaveRanges ports saveRangesForTest.
//
// The diff is scoped to the NUMERIC rows: dictionary limits never enter the map,
// so they are neither updated nor deleted, and a save that mentions none of them
// leaves them exactly where they were. An incoming id that is not among the
// numeric rows — one from another test, or a stale one — is not an error and not
// a match: it is inserted as a NEW row and the id is discarded.
//
// The reporting range and the dictionary normal are NOT editor-managed. An
// existing row keeps whatever it had; a new one keeps the ±Infinity defaults.
func (d *EditorTestDAO) SaveRanges(testID string, desired []DesiredRange, sysUserID int64) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		ts := time.Now().UTC().Truncate(time.Millisecond)

		existing := []RangeRow{}
		if err := tx.Raw(rangeSelect+` WHERE test_id = ? AND test_result_type_id::text = ? ORDER BY id`,
			testID, numericResultTypeID).Scan(&existing).Error; err != nil {
			return err
		}
		numericByID := map[string]RangeRow{}
		for _, r := range existing {
			numericByID[r.ID] = r
		}

		kept := map[string]bool{}
		for _, want := range desired {
			if want.ID != "" {
				if _, ok := numericByID[want.ID]; ok {
					kept[want.ID] = true
					// An UPDATE writes no history row — only inserts and deletes
					// are audited on this table.
					if err := tx.Exec(`
						UPDATE clinlims.result_limits
						   SET component_id = ?, gender = ?, min_age = ?, max_age = ?,
						       low_normal = ?, high_normal = ?, low_critical = ?, high_critical = ?,
						       low_valid = ?, high_valid = ?, lastupdated = ?
						 WHERE id = ?`,
						want.ComponentID, want.Gender, want.MinAge, want.MaxAge,
						want.LowNormal, want.HighNormal, want.LowCritical, want.HighCritical,
						want.LowValid, want.HighValid, ts, want.ID).Error; err != nil {
						return err
					}
					continue
				}
			}

			var newID string
			if err := tx.Raw(`
				INSERT INTO clinlims.result_limits
				       (id, test_id, test_result_type_id, component_id, gender, min_age, max_age,
				        low_normal, high_normal, low_critical, high_critical, low_valid, high_valid,
				        low_reporting_range, high_reporting_range, normal_dictionary_id,
				        always_validate, lastupdated)
				VALUES (nextval('clinlims.result_limits_seq'), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
				        '-Infinity'::float8, 'Infinity'::float8, NULL, false, ?)
				RETURNING id::text`,
				testID, numericResultTypeID, want.ComponentID, want.Gender,
				want.MinAge, want.MaxAge, want.LowNormal, want.HighNormal,
				want.LowCritical, want.HighCritical, want.LowValid, want.HighValid, ts).
				Scan(&newID).Error; err != nil {
				return err
			}
			if err := d.Audit.Write(tx, "RESULT_LIMITS", newID, sysUserID,
				audittrail.ActivityInsert, nil, ts); err != nil {
				return err
			}
		}

		for _, row := range existing {
			if kept[row.ID] {
				continue
			}
			if err := tx.Exec(`DELETE FROM clinlims.result_limits WHERE id = ?`, row.ID).Error; err != nil {
				return err
			}
			changes := rangeDeletePayload(testID, row)
			if err := d.Audit.Write(tx, "RESULT_LIMITS", row.ID, sysUserID,
				audittrail.ActivityDelete, &changes, ts); err != nil {
				return err
			}
		}
		return nil
	})
}

// rangeDeletePayload is getChanges over ResultLimit: the DECLARED-field order,
// and only the fields that differ from a blank object — which is why `gender`
// and `dictionaryNormalId` drop out when they are null.
func rangeDeletePayload(testID string, r RangeRow) string {
	out := audittrail.Field("testId", testID) +
		audittrail.Field("resultTypeId", r.ResultTypeID)
	if r.Gender != nil && *r.Gender != "" {
		out += audittrail.Field("gender", *r.Gender)
	}
	out += audittrail.Field("minAge", rangeDouble(r.MinAge)) +
		audittrail.Field("highCritical", rangeDouble(r.HighCritical)) +
		audittrail.Field("lowCritical", rangeDouble(r.LowCritical)) +
		audittrail.Field("maxAge", rangeDouble(r.MaxAge)) +
		audittrail.Field("lowNormal", rangeDouble(r.LowNormal)) +
		audittrail.Field("highNormal", rangeDouble(r.HighNormal)) +
		audittrail.Field("lowValid", rangeDouble(r.LowValid)) +
		audittrail.Field("highValid", rangeDouble(r.HighValid)) +
		audittrail.Field("lowReportingRange", rangeDouble(r.LowReporting)) +
		audittrail.Field("highReportingRange", rangeDouble(r.HighReporting))
	if r.DictionaryNormalID != nil && *r.DictionaryNormalID != "" {
		out += audittrail.Field("dictionaryNormalId", *r.DictionaryNormalID)
	}
	out += audittrail.Field("alwaysValidate", strconv.FormatBool(r.AlwaysValidate))
	if r.ComponentID != nil && *r.ComponentID != "" {
		out += audittrail.Field("componentId", *r.ComponentID)
	}
	if r.Lastupdated != nil {
		out += audittrail.Field("lastupdated", *r.Lastupdated)
	}
	return out
}

// rangeDouble renders a bound the way Double.toString does — "0.0", "Infinity"
// — without the quotes the JSON form needs.
func rangeDouble(v float64) string {
	s := util.JavaDoubleString(v)
	if math.IsInf(v, 0) || math.IsNaN(v) {
		return s[1 : len(s)-1]
	}
	return s
}
