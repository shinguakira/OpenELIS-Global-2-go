package service

import (
	"sort"
	"strconv"
	"strings"
	"time"

	commonservices "openelis-go/internal/common/services"
	"openelis-go/internal/sample/daoimpl"
	"openelis-go/internal/sample/form"
)

// AbleToCancelRoles is ABLE_TO_CANCEL_ROLE_NAMES from SampleEditRestController.
// The controller resolves the flag from the principal; the list lives here so
// the rule sits next to the code that consumes it.
var AbleToCancelRoles = []string{"Validator", "Validation", "Biologist"}

// SampleEditRequest carries the per-request inputs the form depends on.
//
// They are parameters rather than service fields because all three vary per
// CALLER, and an earlier version of this port pinned each of them to whatever
// the e2e admin happened to produce.
type SampleEditRequest struct {
	AccessionNumber string
	// SysUserID is the AUTHENTICATED user; the role-filtered sampleTypes list
	// depends on their lab-unit roles.
	SysUserID string
	// Editable mirrors isEditable: "readwrite" on the session attribute or on
	// the ?type parameter.
	Editable bool
	// AllowedToCancelAll is isUserAdmin(request) || userInRole(AbleToCancelRoles).
	AllowedToCancelAll bool
}

// SampleEditService backs GET rest/SampleEdit (Wave 4.7).
type SampleEditService struct {
	DAO   *daoimpl.SampleEditDAOImpl
	Lists *commonservices.DisplayListService
	// Status resolves an analysis status id to its DISPLAYED name.
	Status StatusResolver
}

// GetSampleEdit builds the form for an accession number, which may be empty.
//
// Three states, matching Java exactly:
//   - no accessionNumber   -> searchFinished false, noSampleFound false, and
//     accessionNumber itself absent
//   - unknown accession    -> searchFinished true, noSampleFound TRUE, the
//     value echoed back, patient scalars present and empty
//   - found                -> the full form
func (s *SampleEditService) GetSampleEdit(req SampleEditRequest) (*form.SampleEditDTO, error) {
	accessionNumber, sysUserID, editable := req.AccessionNumber, req.SysUserID, req.Editable
	dto := &form.SampleEditDTO{
		FormName:   "SampleEditForm",
		FormAction: "SampleEdit",
		FormMethod: "POST",

		CancelAction: "Home",
		CancelMethod: "POST",

		CurrentDate: currentDateAsText(),
		// Hardcoded true on this form; SamplePatientEntry sets it false.
		Warning: true,

		// isEditable is 'readwrite' on the session attribute OR on the ?type
		// parameter. The session half needs write access this read-only port does
		// not have; the parameter half is honoured.
		IsEditable: editable,
	}

	// Read from site_information rather than hardcoded: the three lengths are
	// derived from the accession prefix, so a deployment with a different prefix
	// reports different numbers.
	acc, err := s.Lists.AccessionConfiguration()
	if err != nil {
		return nil, err
	}
	dto.AccessionFormat = acc.Format
	dto.IDSeparator = acc.IDSeparator
	dto.MaxAccessionLength = acc.MaxAccessionLength
	dto.EditableAccession = acc.EditableAccession
	dto.NonEditableAccession = acc.NonEditableAccession

	conditions, err := s.Lists.InitialSampleConditionList()
	if err != nil {
		return nil, err
	}
	dto.InitialSampleConditionList = conditions

	accessionNumber = strings.TrimSpace(accessionNumber)
	if accessionNumber == "" {
		// The blank form: searchFinished false and accessionNumber never
		// assigned, so NON_NULL drops it. noSampleFound stays false — it is
		// NOT set on this branch, which is why a blank form and an unknown
		// accession differ.
		return dto, nil
	}

	dto.AccessionNumber = &accessionNumber
	dto.SearchFinished = true

	sample, err := s.DAO.SampleByAccession(accessionNumber)
	if err != nil {
		return nil, err
	}
	if sample == nil {
		dto.NoSampleFound = true
		return dto, nil
	}

	if patient, err := s.DAO.PatientForAccession(accessionNumber); err != nil {
		return nil, err
	} else if patient != nil {
		dto.PatientID = patient.PatientID
		// getLastFirstName: "Last, First" — comma AND space.
		dto.PatientName = patient.LastName + ", " + patient.FirstName
		// The stored entered_birth_date, RAW. order/search reformats the same
		// column; this endpoint does not.
		dto.DOB = patient.DOB
		dto.Gender = patient.Gender
		dto.NationalID = patient.NationalID
		dto.SubjectNumber = patient.SubjectNumber
	}

	items, err := s.DAO.EnteredSampleItems(accessionNumber)
	if err != nil {
		return nil, err
	}

	existing, err := s.buildExistingTests(items, accessionNumber, req.AllowedToCancelAll)
	if err != nil {
		return nil, err
	}
	dto.ExistingTests = existing

	possible, err := s.buildPossibleTests(items, accessionNumber)
	if err != nil {
		return nil, err
	}
	dto.PossibleTests = possible

	// The AUTHENTICATED user, not a fixed id: Java passes getSysUserId(request),
	// and the list is filtered by that user's lab-unit roles. Hardcoding one id
	// would serve every caller the admin's sample types.
	sampleTypes, err := s.Lists.UserSampleTypes(sysUserID)
	if err != nil {
		return nil, err
	}
	dto.SampleTypes = sampleTypes

	orderItems, err := s.buildOrderItems(accessionNumber, sample)
	if err != nil {
		return nil, err
	}
	dto.SampleOrderItems = orderItems

	// The suffix is the LAST FILTERED item's sort order, or "-0" when the
	// filter left nothing. Not the sample's highest sort order: E2E-EDIT-02's
	// last item fails the status filter, so this ends "-1" while the table says
	// 2.
	if len(items) > 0 {
		dto.MaxAccessionNumber = accessionNumber + "-" + items[len(items)-1].SortOrder
	} else {
		dto.MaxAccessionNumber = accessionNumber + "-0"
	}

	dto.IsConfirmationSample = sample.IsConfirmation != nil && *sample.IsConfirmation
	// hasResults returns FALSE outright when the caller may not cancel — the
	// flag gates the whole check, it is not just an extra condition.
	dto.AbleToCancelResults = req.AllowedToCancelAll && hasCancellableResults(existing)

	return dto, nil
}

// buildExistingTests ports getCurrentTestInfo + addCurrentTestsToList.
func (s *SampleEditService) buildExistingTests(items []daoimpl.SampleEditItemRow, accessionNumber string, allowedToCancelAll bool) ([]form.SampleEditItemDTO, error) {
	out := []form.SampleEditItemDTO{}
	for _, item := range items {
		analyses, err := s.DAO.AnalysesForItem(item.ID)
		if err != nil {
			return nil, err
		}
		if len(analyses) == 0 {
			// Java appends nothing for an item with no surviving analyses —
			// not even a header row.
			continue
		}

		group := make([]form.SampleEditItemDTO, 0, len(analyses))
		canRemove := true
		for _, a := range analyses {
			// canCancel = allowedToCancelAll
			//             || (status is NOT Canceled AND status IS NotStarted)
			//
			// An earlier version of this port hardcoded it to true "because the
			// e2e user is an admin". That is a test-shaped shortcut: a
			// non-admin without the cancel roles gets a status-dependent answer
			// in Java and got an unconditional true here.
			//
			// The Canceled half is already guaranteed — AnalysesForItem excludes
			// that status — so what remains is the NotStarted check.
			canCancel := allowedToCancelAll ||
				strings.EqualFold(a.StatusName, analysisStatusNotStarted)
			if !canCancel {
				canRemove = false
			}
			// getStatusNameFromId returns the LOCALIZED name, not the DB one:
			// status_of_sample.name "Not Tested" displays as "Not started"
			// through display_key status.test.notStarted. Emitting the column
			// value is a different string.
			status := a.StatusName
			if s.Status != nil {
				if label := s.Status.LabelByName(statusTypeAnalysis, a.StatusName); label != "" {
					status = label
				}
			}
			group = append(group, form.SampleEditItemDTO{
				TestID:       a.TestID,
				TestName:     a.TestName,
				SampleItemID: item.ID,
				SortOrder:    a.TestSortOrder,
				ID:           a.TestID,
				AnalysisID:   &a.AnalysisID,
				Status:       &status,
				CanCancel:    canCancel,
				// hasResults is "the analysis is NOT NotStarted".
				HasResults: !strings.EqualFold(a.StatusName, "Not Tested"),
			})
		}

		sortSampleEditItems(group)

		// Header fields go on the FIRST row of the group only. The rest leave
		// them null and Include.NON_NULL drops them.
		header := accessionNumber + "-" + item.SortOrder
		group[0].AccessionNumber = &header
		if item.TypeOfSample != nil {
			group[0].SampleType = item.TypeOfSample
		}
		group[0].CanRemoveSample = canRemove
		// DateUtil.convertTimestampToStringDate / ...StringTime — the same
		// split order/search uses, so the helper is shared rather than
		// re-derived.
		collectionDate, collectionTime := displayDateTime(item.CollectionDate)
		group[0].CollectionDate = &collectionDate
		group[0].CollectionTime = &collectionTime

		out = append(out, group...)
	}
	return out, nil
}

// buildPossibleTests ports setAddableTestInfo + addPossibleTestsToList.
func (s *SampleEditService) buildPossibleTests(items []daoimpl.SampleEditItemRow, accessionNumber string) ([]form.SampleEditItemDTO, error) {
	out := []form.SampleEditItemDTO{}
	for _, item := range items {
		if item.TypeOfSampleID == nil {
			continue
		}
		tests, err := s.DAO.PossibleTestsForSampleType(*item.TypeOfSampleID)
		if err != nil {
			return nil, err
		}
		if len(tests) == 0 {
			continue
		}

		group := make([]form.SampleEditItemDTO, 0, len(tests))
		for _, t := range tests {
			group = append(group, form.SampleEditItemDTO{
				TestID:       t.TestID,
				TestName:     t.TestName,
				SampleItemID: item.ID,
				SortOrder:    t.SortOrder,
				ID:           t.TestID,
			})
		}
		sortSampleEditItems(group)

		header := accessionNumber + "-" + item.SortOrder
		group[0].AccessionNumber = &header
		if item.TypeOfSample != nil {
			group[0].SampleType = item.TypeOfSample
		}

		out = append(out, group...)
	}
	return out, nil
}

// buildOrderItems assembles the SampleEdit variant of sampleOrderItems.
func (s *SampleEditService) buildOrderItems(accessionNumber string, sample *daoimpl.SampleEditSampleRow) (*form.SampleEditOrderDTO, error) {
	payment, err := s.Lists.PaymentOptions()
	if err != nil {
		return nil, err
	}
	program, err := s.Lists.ProgramList()
	if err != nil {
		return nil, err
	}
	providers, err := s.Lists.ProvidersList()
	if err != nil {
		return nil, err
	}
	referring, err := s.Lists.ReferringSiteList()
	if err != nil {
		return nil, err
	}
	locations, err := s.Lists.TestLocationCodeList()
	if err != nil {
		return nil, err
	}

	priority := "ROUTINE"
	if sample.Priority != nil && *sample.Priority != "" {
		priority = *sample.Priority
	}

	// receivedDateForDisplay / receivedTime are stamped at LOAD TIME from the
	// clock, not read off the sample — getBaseSampleOrderItem sets them before
	// any sample is looked at.
	return &form.SampleEditOrderDTO{
		LabNo:                  accessionNumber,
		SampleID:               sample.ID,
		Priority:               priority,
		ReceivedDateForDisplay: currentDateAsText(),
		ReceivedTime:           currentTimeAsText(),
		PaymentOptions:         payment,
		PriorityList:           s.Lists.PriorityList(),
		ProgramList:            program,
		ProvidersList:          providers,
		ReferringSiteList:      referring,
		TestLocationCodeList:   locations,
		EnvironmentalFields:    map[string]any{},
	}, nil
}

// currentDateAsText ports DateUtil.getCurrentDateAsText — dd/MM/yyyy in the
// display zone, the same zone order/search renders in.
func currentDateAsText() string {
	return time.Now().In(displayZone()).Format("02/01/2006")
}

// currentTimeAsText ports DateUtil.convertTimestampToStringHourTime — HH:mm.
func currentTimeAsText() string {
	return time.Now().In(displayZone()).Format("15:04")
}

// sortSampleEditItems ports SampleEditItemComparator: numeric on sortOrder,
// falling back to a testName comparison when either sortOrder is blank or not
// an integer.
func sortSampleEditItems(items []form.SampleEditItemDTO) {
	sort.SliceStable(items, func(i, j int) bool {
		a, b := items[i], items[j]
		ai, aerr := strconv.Atoi(a.SortOrder)
		bi, berr := strconv.Atoi(b.SortOrder)
		if a.SortOrder == "" || b.SortOrder == "" || aerr != nil || berr != nil {
			return a.TestName < b.TestName
		}
		return ai < bi
	})
}

// hasCancellableResults ports hasResults(currentTestList, allowedToCancelResults).
func hasCancellableResults(items []form.SampleEditItemDTO) bool {
	for _, i := range items {
		if i.HasResults {
			return true
		}
	}
	return false
}
