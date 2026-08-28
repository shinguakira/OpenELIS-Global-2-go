// Package daoimpl ports the org.openelisglobal.sample DAO reads behind the c2
// endpoints. Folder layout mirrors the Java source during migration.
package daoimpl

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"openelis-go/internal/sample/valueholder"
)

// SampleDAOImpl ports SampleDAOImpl for the c2 read paths.
type SampleDAOImpl struct {
	DB *gorm.DB
}

// ErrAttachmentNotFound signals an attachment id with no row at all, as opposed
// to one that exists and is soft-deleted. Java distinguishes the two by
// accident — OrderAttachmentServiceImpl.get throws for a missing id while
// serveAttachment's own guard handles the deleted one — and the two produce
// DIFFERENT statuses (500 vs 404), so the distinction has to survive the port.
var ErrAttachmentNotFound = errors.New("order attachment not found")

// GetByAccessionNumber mirrors SampleServiceImpl.getSampleByAccessionNumber.
//
// Java strips anything from the first '.' onward before querying
// (SampleServiceImpl: `if (labNumber.contains(".")) labNumber =
// labNumber.substring(0, labNumber.indexOf('.'))`), so "E2E001.1" and "E2E001"
// resolve to the same sample. Reproduced in the service, not here, so this DAO
// stays a plain lookup like Java's.
//
// Returns (nil, nil) on no match — the controller turns that into Java's 404.
func (d *SampleDAOImpl) GetByAccessionNumber(accessionNumber string) (*valueholder.Sample, error) {
	var s valueholder.Sample
	err := d.DB.Where("accession_number = ?", accessionNumber).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// AnalysisRow is one row of the join behind convertToForm.
type AnalysisRow struct {
	AnalysisID   int64   `gorm:"column:analysis_id"`
	SampleType   *string `gorm:"column:sample_type"`
	ReferralTest *string `gorm:"column:referral_test"`
}

// AnalysesForSampleByStatus mirrors
// AnalysisDAOImpl.getAnalysesBySampleIdAndStatusId plus the two lookups
// SampleRestController.convertToForm performs per analysis
// (analysis.getSampleItem().getTypeOfSample().getDescription() and
// analysis.getTest().getName()), folded into one query.
//
// Both are LEFT joins because Java null-guards both before setting the field;
// an inner join would silently drop analyses whose test or sample type is
// unset, shortening the list.
//
// Note it is `description`, not `name` or `local_abbrev`, that feeds
// sampleType — Java calls getDescription().
//
// ORDERING: Java's HQL has no ORDER BY, so its order is DB-natural and
// undefined. The explicit ORDER BY here is a divergence in shape only — it
// makes the response deterministic without changing which rows appear, and the
// e2e oracle compares row SETS, never positions.
func (d *SampleDAOImpl) AnalysesForSampleByStatus(sampleID int64, statusIDs []string) ([]AnalysisRow, error) {
	rows := []AnalysisRow{}
	q := d.DB.Table("clinlims.analysis AS a").
		Select("a.id AS analysis_id, tos.description AS sample_type, t.name AS referral_test").
		Joins("JOIN clinlims.sample_item AS si ON si.id = a.sampitem_id").
		Joins("LEFT JOIN clinlims.type_of_sample AS tos ON tos.id = si.typeosamp_id").
		Joins("LEFT JOIN clinlims.test AS t ON t.id = a.test_id").
		Where("si.samp_id = ?", sampleID)
	if len(statusIDs) > 0 {
		q = q.Where("a.status_id IN (?)", statusIDs)
	}
	if err := q.Order("a.id ASC").Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// PendingAnalysisRow is one entry of the getPendingAnalysisForTestProvider
// groups: the accession the analysis belongs to, and the analysis id.
type PendingAnalysisRow struct {
	LabNo      string `gorm:"column:lab_no"`
	AnalysisID int64  `gorm:"column:analysis_id"`
}

// AnalysesByTestAndStatus mirrors
// AnalysisDAOImpl.getAllAnalysisByTestAndStatus:
//
//	from Analysis a where a.test.id = :testId and a.statusId IN (:statusIdList)
//	order by a.sampleItem.sample.accessionNumber
//
// The ORDER BY is Java's own, not an addition — the response arrays are ordered
// by accession, and a port that returned DB-natural order would diverge
// visibly on any test with more than one pending analysis.
//
// The joins are inner because the controller dereferences
// analysis.getSampleItem().getSample().getAccessionNumber() unguarded: an
// analysis whose chain is broken would NPE in Java, not be skipped, so there is
// no "keep the row" behavior to preserve.
func (d *SampleDAOImpl) AnalysesByTestAndStatus(testID string, statusIDs []string) ([]PendingAnalysisRow, error) {
	rows := []PendingAnalysisRow{}
	if len(statusIDs) == 0 {
		return rows, nil
	}
	err := d.DB.Table("clinlims.analysis AS a").
		Select("s.accession_number AS lab_no, a.id AS analysis_id").
		Joins("JOIN clinlims.sample_item AS si ON si.id = a.sampitem_id").
		Joins("JOIN clinlims.sample AS s ON s.id = si.samp_id").
		Where("a.test_id = ? AND a.status_id IN (?)", testID, statusIDs).
		Order("s.accession_number ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// OrderAttachmentRow is one clinlims.order_attachment row, minus file_content —
// the blob is never exposed by the list endpoint (it has its own
// download/view routes) and pulling it would move megabytes per request.
type OrderAttachmentRow struct {
	ID            int64      `gorm:"column:id"`
	FileName      string     `gorm:"column:original_file_name"`
	FileType      *string    `gorm:"column:file_type"`
	FileSizeBytes *int64     `gorm:"column:file_size_bytes"`
	UploadedAt    *time.Time `gorm:"column:uploaded_at"`
}

// OrderAttachmentBlob is one attachment WITH its bytes, for the download/view
// routes. Separate from OrderAttachmentRow because the list endpoint must never
// pull file_content.
//
// IsDeleted is carried rather than filtered in SQL: serveAttachment fetches by
// id and then refuses a soft-deleted row, which is what makes a deleted id a
// 404 while a MISSING id is a 500 (see AttachmentByID).
type OrderAttachmentBlob struct {
	ID          int64   `gorm:"column:id"`
	FileName    string  `gorm:"column:original_file_name"`
	FileType    *string `gorm:"column:file_type"`
	FileContent []byte  `gorm:"column:file_content"`
	IsDeleted   bool    `gorm:"column:is_deleted"`
}

// AttachmentByID mirrors OrderAttachmentServiceImpl.get(attachmentId).
//
// Returns (nil, nil) for an id that exists but is soft-deleted? No — it returns
// the row and lets the caller decide, matching Java. For an id that does not
// exist at all it returns ErrAttachmentNotFound, because Java's get() THROWS
// there rather than returning null: that is why a missing id is a 500 while a
// soft-deleted one is a 404. Two "not there" conditions, two statuses; pinned,
// not smoothed over.
func (d *SampleDAOImpl) AttachmentByID(id int64) (*OrderAttachmentBlob, error) {
	rows := []OrderAttachmentBlob{}
	err := d.DB.Table("clinlims.order_attachment").
		Select("id, original_file_name, file_type, file_content, is_deleted").
		Where("id = ?", id).
		Limit(1).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, ErrAttachmentNotFound
	}
	return &rows[0], nil
}

// ActiveAttachmentsBySampleID mirrors
// OrderAttachmentDAOImpl.findActiveBySampleId:
//
//	from OrderAttachment oa where oa.sampleId = :sampleId
//	and oa.isDeleted = false order by oa.uploadedAt desc
//
// The ORDER BY is Java's, and it is DESC — newest first. Soft-deleted rows are
// excluded here rather than in the service, matching where Java puts it.
func (d *SampleDAOImpl) ActiveAttachmentsBySampleID(sampleID int64) ([]OrderAttachmentRow, error) {
	rows := []OrderAttachmentRow{}
	err := d.DB.Table("clinlims.order_attachment").
		Select("id, original_file_name, file_type, file_size_bytes, uploaded_at").
		Where("sample_id = ? AND is_deleted = false", sampleID).
		Order("uploaded_at DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}
