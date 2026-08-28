package service

import (
	"errors"
	"strconv"
	"time"

	"openelis-go/internal/sample/daoimpl"
	"openelis-go/internal/sample/form"
)

// UnassignedSampleService ports UnassignedSampleServiceImpl and
// UnassignedSampleItemServiceImpl — the read half of the shipment
// unassigned-sample dashboard.
//
// Row shapes here are compared against live Java, not read off the Java source:
// src/test/resources/fixtures/shipment-attachment-e2e.sql seeds ten referrals —
// five visible and five excluded, one per exclusion rule — because
// clinlims.referral is empty in the stock dataset and every method here used to
// return an empty result on both stacks. That gap hid two real defects: this
// file emitted referralDate as a formatted string where Java emits epoch
// millis, and it carried an invented DTO for /items that Java never produces.
type UnassignedSampleService struct {
	DAO *daoimpl.UnassignedSampleDAOImpl
}

// GetUnassignedForDashboard mirrors getUnassignedSamplesForDashboard: every
// unassigned referral, each run through compileSampleData.
func (s *UnassignedSampleService) GetUnassignedForDashboard() ([]form.UnassignedSampleDTO, error) {
	rows, err := s.DAO.UnassignedReferrals()
	if err != nil {
		return nil, err
	}
	return compileSampleData(rows, nil), nil
}

// GetUnassignedByFacility mirrors getUnassignedSamplesByDestinationFacility.
//
// Java filters IN JAVA, not in SQL: it loads every unassigned referral and
// keeps those whose organization id string-equals the requested one. Referrals
// with no organization are dropped, because the guard is
// `referral.getOrganization() != null && ...`.
func (s *UnassignedSampleService) GetUnassignedByFacility(facilityID int64) ([]form.UnassignedSampleDTO, error) {
	rows, err := s.DAO.UnassignedReferrals()
	if err != nil {
		return nil, err
	}
	return compileSampleData(rows, &facilityID), nil
}

// CountUnassignedByFacility mirrors countUnassignedSamplesByFacility.
//
// NOT simply len(GetUnassignedByFacility). Java applies an EXTRA filter here
// that the listing method does not:
//
//	if (Boolean.TRUE.equals(referral.getLostStatus()) || referral.isCanceled()) continue;
//
// so in principle the count could be a SUBSET of the list length. Measured
// against live Java with the referral fixture loaded, the two are EQUAL — the
// extra pass removes nothing, because getUnassignedReferrals has already
// excluded lost and canceled in SQL. The c2 spec asserts that equality rather
// than the weaker `count <= length` an earlier draft used, which any
// implementation returning 0 forever would have satisfied.
//
// The redundant filter is kept anyway: it is Java's, and dropping it would make
// this port depend on that SQL redundancy holding forever.
func (s *UnassignedSampleService) CountUnassignedByFacility(facilityID int64) (int, error) {
	rows, err := s.DAO.UnassignedReferrals()
	if err != nil {
		return 0, err
	}
	count := 0
	for _, r := range rows {
		if r.DestinationFacilityID != nil && *r.DestinationFacilityID == facilityID {
			count++
		}
	}
	return count, nil
}

// ErrUnassignedItemsUnserializable reproduces a JAVA DEFECT, deliberately.
//
// MIGRATION POLICY: pin Java, do not fix it. Raised separately for the
// maintainers rather than silently corrected here — a port that "helpfully"
// returns the list makes Go and Java disagree on a live endpoint, which is the
// one thing this migration must not do.
//
// UnassignedSampleItemServiceImpl.buildSampleItemDTOs calls
// referralDAO.getReferralsBySampleItemId(Integer) once per result row, but
// SampleItem.id is mapped as a String, so Hibernate refuses the binding:
//
//	IllegalArgumentException: Parameter value [10111] did not match
//	expected type [java.lang.String (n/a)]
//
// The service catches it and returns an empty list, but the exception has
// already marked the read-only transaction rollback-only, so the commit at the
// @Transactional boundary throws and the CONTROLLER's catch answers
// ResponseEntity.status(500).build() — 500 with Content-Length: 0.
//
// The failure is structural (Integer argument vs String-mapped id), not
// value-dependent, so it fires for ANY non-empty result. A query that matches
// nothing never enters the loop and still answers 200 [].
var ErrUnassignedItemsUnserializable = errors.New(
	"java parity: unassigned-sample/items throws once any row matches (Integer bound to String-mapped SampleItem.id)")

// GetUnassignedItems mirrors UnassignedSampleItemServiceImpl.getAllUnassigned
// and, with a non-empty accessionNumber, searchUnassignedByAccessionNumber —
// the two differ only by a LIKE clause.
//
// Returns ErrUnassignedItemsUnserializable when the query matches anything, and
// an empty list otherwise. Both branches are measured against live Java; see
// the c2 spec's "500 once ANY row matches" test.
//
// The DTO-building that used to live here is gone. It produced a shape Java has
// never emitted — invented field names included — and nothing could catch that
// while clinlims.referral was empty and both stacks answered 200 [].
func (s *UnassignedSampleService) GetUnassignedItems(accessionNumber string) ([]any, error) {
	excluded, err := s.DAO.AssignedSampleItemIDs()
	if err != nil {
		return nil, err
	}
	rows, err := s.DAO.UnassignedSampleItems(daoimpl.UnassignedSampleItemsQuery{
		AccessionNumber:       accessionNumber,
		ExcludedSampleItemIDs: excluded,
	})
	if err != nil {
		return nil, err
	}
	if len(rows) > 0 {
		return nil, ErrUnassignedItemsUnserializable
	}
	return []any{}, nil
}

// compileSampleData ports UnassignedSampleServiceImpl.compileSampleData. When
// facilityID is non-nil the caller wants only that destination facility, which
// Java applies as an in-Java filter before compiling.
func compileSampleData(rows []daoimpl.UnassignedReferralRow, facilityID *int64) []form.UnassignedSampleDTO {
	out := make([]form.UnassignedSampleDTO, 0, len(rows))
	for _, r := range rows {
		if facilityID != nil {
			if r.DestinationFacilityID == nil || *r.DestinationFacilityID != *facilityID {
				continue
			}
		}
		dto := form.UnassignedSampleDTO{
			ID: strconv.FormatInt(r.ID, 10),
			// Java defaults a null priority to the literal "Normal".
			Priority: "Normal",
		}
		if r.Priority != nil && *r.Priority != "" {
			dto.Priority = *r.Priority
		}
		if r.ReferralDate != nil {
			// Epoch millis, NOT a formatted string: the raw Timestamp goes
			// into the HashMap and Jackson writes it as a number. See the
			// field comment on UnassignedSampleDTO.ReferralDate.
			ms := r.ReferralDate.UnixMilli()
			dto.ReferralDate = &ms
			// TimeUnit.MILLISECONDS.toDays truncates toward zero, so a
			// referral raised today reports 0 rather than rounding up.
			dto.DaysUnassigned = int64(time.Since(*r.ReferralDate) / (24 * time.Hour))
		}
		if r.AccessionNumber != nil {
			dto.AccessionNumber = r.AccessionNumber
		}
		if r.SampleID != nil {
			id := strconv.FormatInt(*r.SampleID, 10)
			dto.SampleID = &id
		}
		if r.ReferralTestName != nil {
			dto.ReferralTestName = r.ReferralTestName
		}
		if r.TestID != nil {
			id := strconv.FormatInt(*r.TestID, 10)
			dto.TestID = &id
		}
		// Java prefers the joined organization's name and falls back to the
		// referral's own organization_name column ONLY when there is no
		// organization at all — and in that fallback it sets the name without
		// an id, so the two keys do not always travel together.
		if r.DestinationFacilityID != nil {
			id := strconv.FormatInt(*r.DestinationFacilityID, 10)
			dto.DestinationFacilityID = &id
			dto.DestinationFacilityName = r.DestinationFacilityName
		} else if r.OrganizationName != nil {
			dto.DestinationFacilityName = r.OrganizationName
		}
		if r.ReferralReasonID != nil {
			id := strconv.FormatInt(*r.ReferralReasonID, 10)
			dto.ReferralReasonID = &id
		}
		out = append(out, dto)
	}
	return out
}
