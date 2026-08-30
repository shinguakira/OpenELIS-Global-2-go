package daoimpl

import (
	"strconv"
	"time"

	"gorm.io/gorm"

	"openelis-go/internal/common/audittrail"
)

// The TEST audit rows the editor leaves behind.
//
// MEASURED, and it is the one place in this package where clinlims.history is
// written: creating a test leaves an 'I' with a NULL payload, and Basic Info
// and activation leave a 'U' carrying the values they REPLACED. Everything else
// the editor writes — components, options, interpretations, storage,
// terminology, panel membership, display order — is silent.
//
// A save that changes nothing writes nothing. Measured: re-sending the same
// name, and deactivating a test that was already inactive, both answer 200 with
// no history row.

// testAuditFields is the payload's field order: Test's DECLARED-field order,
// with the superclass `isActive` after all of them.
//
// The order is not cosmetic — getChanges walks the reflected field list, so a
// port that emits its own order produces a payload no Java reader parses the
// same way. `testSection` is the odd one: it renders the section's
// DESCRIPTION ("Bacteria logbook"), not its id or its name.
type testAuditSnapshot struct {
	TestSectionDescription *string `gorm:"column:test_section_description"`
	Description            *string `gorm:"column:description"`
	Domain                 *string `gorm:"column:domain"`
	LocalCode              *string `gorm:"column:local_code"`
	Orderable              *bool   `gorm:"column:orderable"`
	AMR                    *bool   `gorm:"column:amr"`
	IsActive               *string `gorm:"column:is_active"`
}

func testAuditState(tx *gorm.DB, testID string) (*testAuditSnapshot, error) {
	rows := []testAuditSnapshot{}
	if err := tx.Raw(`
		SELECT ts.description AS test_section_description, t.description AS description,
		       t.domain AS domain, t.local_code AS local_code, t.orderable AS orderable,
		       t.antimicrobial_resistance AS amr, t.is_active AS is_active
		  FROM clinlims.test t
		  LEFT JOIN clinlims.test_section ts ON ts.id = t.test_section_id
		 WHERE t.id = ?`, testID).Scan(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

// writeTestUpdateAudit compares the state before and after and writes one 'U'
// row carrying the OLD value of every field that moved. No movement, no row.
func writeTestUpdateAudit(tx *gorm.DB, audit *audittrail.Service, testID string,
	before, after *testAuditSnapshot, sysUserID int64, ts time.Time) error {

	if before == nil || after == nil {
		return nil
	}
	changes := ""
	if differs(before.TestSectionDescription, after.TestSectionDescription) {
		changes += audittrail.Field("testSection", deref(before.TestSectionDescription))
	}
	if differs(before.Description, after.Description) {
		changes += audittrail.Field("description", deref(before.Description))
	}
	if differs(before.Domain, after.Domain) {
		changes += audittrail.Field("domain", deref(before.Domain))
	}
	if differs(before.LocalCode, after.LocalCode) {
		changes += audittrail.Field("localCode", deref(before.LocalCode))
	}
	if differsBool(before.Orderable, after.Orderable) {
		changes += audittrail.Field("orderable", boolText(before.Orderable))
	}
	if differsBool(before.AMR, after.AMR) {
		changes += audittrail.Field("antimicrobialResistance", boolText(before.AMR))
	}
	if differs(before.IsActive, after.IsActive) {
		changes += audittrail.Field("isActive", deref(before.IsActive))
	}
	if changes == "" {
		return nil
	}
	return audit.Write(tx, "TEST", testID, sysUserID, audittrail.ActivityUpdate, &changes, ts)
}

func differs(a, b *string) bool {
	if a == nil || b == nil {
		return (a == nil) != (b == nil)
	}
	return *a != *b
}

func differsBool(a, b *bool) bool {
	if a == nil || b == nil {
		return (a == nil) != (b == nil)
	}
	return *a != *b
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func boolText(b *bool) string {
	if b == nil {
		return ""
	}
	return strconv.FormatBool(*b)
}
