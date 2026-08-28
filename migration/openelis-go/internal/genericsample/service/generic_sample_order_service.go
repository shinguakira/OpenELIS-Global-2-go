// Package service ports
// org.openelisglobal.genericsample.service.GenericSampleOrderService for the
// Wave 4.5 read. Folder layout mirrors the Java source during migration.
package service

import (
	"gorm.io/gorm"
)

// GenericSampleOrderRollbackMessage is the EXACT body Java returns for an
// accession that exists.
//
// MIGRATION POLICY: this reproduces a JAVA DEFECT rather than fixing it, and it
// is recorded in migration/java-defects-found.md so the maintainers can be told.
//
// getGenericSampleOrderByAccessionNumber reaches a query that binds a numeric
// sample-item id to a String-mapped property:
//
//	Failed to retrieve notebook sample for accession: E2E001
//	Parameter value [10002] did not match expected type [java.lang.String (n/a)]
//
// The exception marks the transaction rollback-only; the commit at the
// @Transactional boundary then throws UnexpectedRollbackException, and the
// handler's catch-all wraps ITS message — not the original Hibernate one — into
// the {error} envelope with a 500. That is why the text below mentions the
// rollback rather than the parameter mismatch.
//
// The failure is structural (numeric argument against a String-mapped id), not
// value-dependent, so every existing accession hits it. Same root cause as
// rest/unassigned-sample/items.
const GenericSampleOrderRollbackMessage = "Failed to retrieve generic sample order: " +
	"Transaction silently rolled back because it has been marked as rollback-only"

// GenericSampleOrderService backs GET rest/GenericSampleOrder.
//
// It needs only to tell "no such accession" from "an accession that exists",
// because those are Java's only two outcomes. Building the form would be dead
// code: Java never serializes one from this path.
type GenericSampleOrderService struct {
	DB *gorm.DB
}

// SampleExists reports whether the accession resolves to a sample.
//
// Java looks the sample up with the same truncating accession lookup the rest of
// c2 uses, but the truncation is NOT applied here: the Java service passes the
// raw parameter through to its own finder. Matching that exactly matters for the
// 404-vs-500 split, which is the only observable this endpoint has.
func (s *GenericSampleOrderService) SampleExists(accessionNumber string) (bool, error) {
	var n int64
	err := s.DB.Table("clinlims.sample").
		Where("accession_number = ?", accessionNumber).
		Count(&n).Error
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
