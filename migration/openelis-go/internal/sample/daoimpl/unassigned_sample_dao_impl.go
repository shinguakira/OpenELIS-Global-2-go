package daoimpl

import (
	"time"

	"gorm.io/gorm"
)

// UnassignedSampleDAOImpl ports ReferralDAOImpl's unassigned-referral reads.
//
// COVERAGE LIMIT, stated up front: clinlims.referral is EMPTY in the dev
// dataset, so every query here returns zero rows and the row-level shape below
// is implemented from the Java source rather than verified against a live
// response. What IS verified is the envelope — array vs object, status codes,
// and the count/list relationship. Seeding a referral fixture is what would
// close the gap.
type UnassignedSampleDAOImpl struct {
	DB *gorm.DB
}

// unassignedReferralPredicate is ReferralDAOImpl.getUnassignedReferrals' WHERE
// clause, shared by every query in this file because Java repeats it verbatim:
//
//	r.assignedBox IS NULL
//	AND (r.lostStatus IS NULL OR r.lostStatus = false)
//	AND r.status != 'CANCELED'
//
// Note it filters on `status != 'CANCELED'`, NOT on the referral.canceled
// column. That column exists in the DB but is not mapped on the entity — which
// is precisely why the sibling getUnassignedSampleByAccessionNumber, the one
// query that DOES reference r.canceled, fails to parse and 500s.
const unassignedReferralPredicate = "r.assigned_to_box_id IS NULL" +
	" AND (r.lost_status IS NULL OR r.lost_status = false)" +
	" AND r.status <> 'CANCELED'"

// UnassignedReferralRow is one row of compileSampleData's inputs.
type UnassignedReferralRow struct {
	ID                      int64      `gorm:"column:id"`
	ReferralDate            *time.Time `gorm:"column:referral_request_date"`
	Priority                *string    `gorm:"column:priority"`
	AccessionNumber         *string    `gorm:"column:accession_number"`
	SampleID                *int64     `gorm:"column:sample_id"`
	ReferralTestName        *string    `gorm:"column:referral_test_name"`
	TestID                  *int64     `gorm:"column:test_id"`
	DestinationFacilityID   *int64     `gorm:"column:destination_facility_id"`
	DestinationFacilityName *string    `gorm:"column:destination_facility_name"`
	OrganizationName        *string    `gorm:"column:organization_name"`
	ReferralReasonID        *int64     `gorm:"column:referral_reason_id"`
}

// UnassignedReferrals mirrors getUnassignedReferrals joined to everything
// compileSampleData dereferences. The joins are LEFT because compileSampleData
// null-guards each hop (analysis -> sampleItem -> sample, test, organization).
func (d *UnassignedSampleDAOImpl) UnassignedReferrals() ([]UnassignedReferralRow, error) {
	rows := []UnassignedReferralRow{}
	err := d.DB.Table("clinlims.referral AS r").
		Select(`r.id, r.referral_request_date, r.priority, r.organization_name,
			r.referral_reason_id,
			s.accession_number AS accession_number, s.id AS sample_id,
			t.name AS referral_test_name, t.id AS test_id,
			o.id AS destination_facility_id, o.name AS destination_facility_name`).
		Joins("LEFT JOIN clinlims.analysis AS a ON a.id = r.analysis_id").
		Joins("LEFT JOIN clinlims.sample_item AS si ON si.id = a.sampitem_id").
		Joins("LEFT JOIN clinlims.sample AS s ON s.id = si.samp_id").
		Joins("LEFT JOIN clinlims.test AS t ON t.id = a.test_id").
		Joins("LEFT JOIN clinlims.organization AS o ON o.id = r.organization_id").
		Where(unassignedReferralPredicate).
		// ORDER BY ctid — the physical order Java's unordered query returns.
		//
		// getUnassignedReferrals is `FROM Referral r WHERE ...` with no ordering
		// at all, so the array order is whatever Postgres scans. This query adds
		// five JOINs that Hibernate resolves lazily per row, and those let the
		// planner emit a different sequence — measured: Java led with E2E-REF-01
		// and this port with E2E-REF-03. The rows are rendered in array order in
		// the shipment dashboard, so that is observable, and every assertion on
		// this endpoint matched by id and never looked at it.
		Order("r.ctid").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// UnassignedSampleItemRow is one row of
// getUnassignedReferralsGroupedBySampleItem's six-column projection.
type UnassignedSampleItemRow struct {
	SampleItemID     int64      `gorm:"column:sample_item_id"`
	AccessionNumber  string     `gorm:"column:accession_number"`
	TypeOfSampleName string     `gorm:"column:type_of_sample_name"`
	TypeOfSampleID   *int64     `gorm:"column:type_of_sample_id"`
	CollectionDate   *time.Time `gorm:"column:collection_date"`
	SortOrder        *string    `gorm:"column:sort_order"`
}

// UnassignedSampleItems mirrors getUnassignedReferralsGroupedBySampleItem, and
// with a non-empty accessionNumber also its searchUnassignedByAccessionNumber
// sibling — the two HQL strings are identical apart from that one LIKE clause.
//
// COALESCE(tos.description, ”) is Java's, so an unmapped sample type comes
// back as an empty string rather than null. ORDER BY accession, sortOrder is
// Java's too.
//
// The excluded-sample-item list (items already in a shipment box) is passed in
// rather than joined, matching how Java assembles it from a separate DAO call.
func (d *UnassignedSampleItemsQuery) run(db *gorm.DB) ([]UnassignedSampleItemRow, error) {
	rows := []UnassignedSampleItemRow{}
	q := db.Table("clinlims.referral AS r").
		Distinct(`si.id AS sample_item_id, s.accession_number AS accession_number,
			COALESCE(tos.description, '') AS type_of_sample_name, tos.id AS type_of_sample_id,
			s.collection_date AS collection_date, si.sort_order AS sort_order`).
		Joins("JOIN clinlims.analysis AS a ON a.id = r.analysis_id").
		Joins("JOIN clinlims.sample_item AS si ON si.id = a.sampitem_id").
		Joins("JOIN clinlims.sample AS s ON s.id = si.samp_id").
		Joins("LEFT JOIN clinlims.type_of_sample AS tos ON tos.id = si.typeosamp_id").
		Where(unassignedReferralPredicate).
		Where("(si.rejected IS NULL OR si.rejected = false)").
		Where("(si.voided IS NULL OR si.voided = false)")
	if q2 := d.AccessionNumber; q2 != "" {
		q = q.Where("s.accession_number LIKE ?", "%"+q2+"%")
	}
	if len(d.ExcludedSampleItemIDs) > 0 {
		q = q.Where("si.id NOT IN (?)", d.ExcludedSampleItemIDs)
	}
	if err := q.Order("s.accession_number, si.sort_order").Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// UnassignedSampleItemsQuery carries the two optional filters.
type UnassignedSampleItemsQuery struct {
	AccessionNumber       string
	ExcludedSampleItemIDs []int64
}

// UnassignedSampleItems runs the grouped-by-sample-item query.
func (d *UnassignedSampleDAOImpl) UnassignedSampleItems(q UnassignedSampleItemsQuery) ([]UnassignedSampleItemRow, error) {
	return q.run(d.DB)
}

// AssignedSampleItemIDs mirrors boxSampleItemDAO.getAllAssignedSampleItemIds —
// sample items already placed in a shipment box, which both item queries
// exclude.
func (d *UnassignedSampleDAOImpl) AssignedSampleItemIDs() ([]int64, error) {
	ids := []int64{}
	err := d.DB.Table("clinlims.box_sample_item").Distinct().Pluck("sample_item_id", &ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}
