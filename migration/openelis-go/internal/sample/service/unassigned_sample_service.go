package service

import (
	"strconv"
	"time"

	"openelis-go/internal/sample/daoimpl"
	"openelis-go/internal/sample/form"
)

// UnassignedSampleService ports UnassignedSampleServiceImpl and
// UnassignedSampleItemServiceImpl — the read half of the shipment
// unassigned-sample dashboard.
//
// COVERAGE LIMIT: clinlims.referral is empty in the dev dataset, so every
// method here returns an empty result and the ROW shape is implemented from the
// Java source without a live response to compare against. The envelope (array
// vs object), the status codes and the count-vs-list relationship ARE verified.
// Seeding a referral fixture is what would close this.
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
// so the count is legitimately a SUBSET of the list length. The c2 spec asserts
// the inequality rather than equality for exactly this reason. Both filters are
// redundant with the SQL predicate today — getUnassignedReferrals already
// excludes lost and canceled — but they are Java's, and a port that dropped
// them would be relying on that redundancy holding forever.
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

// GetUnassignedItems mirrors UnassignedSampleItemServiceImpl.getAllUnassigned
// and, with a non-empty accessionNumber, searchUnassignedByAccessionNumber —
// the two differ only by a LIKE clause.
func (s *UnassignedSampleService) GetUnassignedItems(accessionNumber string) ([]form.UnassignedSampleItemDTO, error) {
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

	// buildSampleItemDTOs makes a first pass counting how many sample items
	// share each accession, then suffixes the DISPLAYED accession with the sort
	// order only when there is more than one. Reproduced, including the "1"
	// default when sort_order is null.
	perAccession := map[string]int{}
	for _, r := range rows {
		perAccession[r.AccessionNumber]++
	}

	dtos := make([]form.UnassignedSampleItemDTO, 0, len(rows))
	for _, r := range rows {
		display := r.AccessionNumber
		if perAccession[r.AccessionNumber] > 1 {
			sortOrder := "1"
			if r.SortOrder != nil && *r.SortOrder != "" {
				sortOrder = *r.SortOrder
			}
			display = r.AccessionNumber + "-" + sortOrder
		}
		dto := form.UnassignedSampleItemDTO{
			ID:                    strconv.FormatInt(r.SampleItemID, 10),
			SampleAccessionNumber: display,
			SampleType:            r.TypeOfSampleName,
			// Java always initialises these to empty lists, so they serialize
			// as [] rather than being dropped by Include.NON_NULL.
			ChildAliquots: []any{},
			OrderedTests:  []any{},
			ReferralTests: []any{},
		}
		if r.TypeOfSampleID != nil {
			id := strconv.FormatInt(*r.TypeOfSampleID, 10)
			dto.SampleTypeID = &id
		}
		if r.CollectionDate != nil {
			ts := formatSQLTimestamp(*r.CollectionDate)
			dto.CollectionDate = &ts
		}
		dtos = append(dtos, dto)
	}
	return dtos, nil
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
			ts := formatSQLTimestamp(*r.ReferralDate)
			dto.ReferralDate = &ts
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
