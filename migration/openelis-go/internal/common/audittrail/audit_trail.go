// Package audittrail ports org.openelisglobal.audittrail.daoimpl.AuditTrailServiceImpl
// — the clinlims.history rows every write leaves behind.
//
// It is invisible from the outside. No response mentions it, so a port that
// wrote only the target row passed every response-level check this project had
// until an e2e spec started reading `history` directly. The wave that found it
// (e1) had already reported "書き込み一致" from a probe that compared only
// site_information.
//
// Java gates the whole mechanism on reference_tables.keep_history: a table
// flagged 'N' is not audited and the write proceeds silently. A table with no
// reference_tables row at all is an ERROR that rolls the write back — the audit
// is not best-effort, and a write that cannot be recorded does not happen.
package audittrail

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// The activity codes IActionConstants defines.
const (
	ActivityInsert = "I"
	ActivityUpdate = "U"
	ActivityDelete = "D"
)

// Service writes history rows.
type Service struct{}

// referenceTable resolves a table name to its reference_tables id and whether
// history is kept for it.
func referenceTable(tx *gorm.DB, tableName string) (id int64, keep bool, err error) {
	var row struct {
		ID          int64  `gorm:"column:id"`
		KeepHistory string `gorm:"column:keep_history"`
	}
	err = tx.Table("clinlims.reference_tables").
		Select("id, keep_history").
		Where("name = ?", tableName).
		Take(&row).Error
	if err != nil {
		return 0, false, err
	}
	return row.ID, row.KeepHistory == "Y", nil
}

// Write inserts one history row.
//
// changes is the audit PAYLOAD, and it is the row's state BEFORE the write, not
// after — an update records the value it replaced. Pass nil for an insert:
// saveNewHistory never sets the column, so it is NULL rather than empty.
func (s *Service) Write(tx *gorm.DB, tableName, referenceID string, sysUserID int64,
	activity string, changes *string, ts time.Time) error {

	tableID, keep, err := referenceTable(tx, tableName)
	if err != nil {
		return fmt.Errorf("audittrail: reference table %q: %w", tableName, err)
	}
	if !keep {
		// keep_history = 'N': not an error, just not audited.
		return nil
	}

	var payload any
	if changes != nil {
		payload = []byte(*changes)
	}

	return tx.Exec(`
		INSERT INTO clinlims.history
		       (id, sys_user_id, reference_id, reference_table, timestamp, activity, changes)
		VALUES (nextval('clinlims.history_seq'), ?, ?, ?, ?, ?, ?)`,
		sysUserID, referenceID, tableID, ts, activity, payload).Error
}

// Field renders one line of the payload: `<label>value</label>` plus a newline.
//
// getXMLFormat emits the tag even when the value is blank — the field name is
// itself the record of what changed — and trims the value before encoding.
func Field(label, value string) string {
	return "<" + label + ">" + escapeXML(strings.TrimSpace(value)) + "</" + label + ">\n"
}

// escapeXML is Encode.forXmlContent: the three characters that cannot appear
// literally in element content.
func escapeXML(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(s)
}

// JavaTimestamp renders a time the way java.sql.Timestamp.toString does, which
// is the format the payload's <lastupdated> carries.
//
// Java writes these timestamps from new Timestamp(System.currentTimeMillis()),
// so the stored value only ever has MILLISECOND precision and toString prints
// exactly three fractional digits for it. A port that let the database supply
// now() would store microseconds and render six.
func JavaTimestamp(t time.Time) string {
	return t.UTC().Format("2006-01-02 15:04:05.000")
}
