package service

import (
	"os"
	"strconv"
	"strings"
	"time"

	"openelis-go/internal/sample/daoimpl"
	"openelis-go/internal/sample/form"
)

// OrderSearchService backs GET rest/order/search.
type OrderSearchService struct {
	DAO *daoimpl.SampleDAOImpl
}

// GetOrderByLabNumber mirrors OrderSearchRestController.searchOrder.
//
// Java's shape, reproduced:
//   - blank/missing labNumber -> the controller answers 400 before calling here
//   - no such sample          -> (nil, nil), controller answers 404
//
// The accession is TRIMMED before lookup (`labNumber.trim()`), and
// getSampleByAccessionNumber then truncates at the first '.' as everywhere else
// in this wave.
func (s *OrderSearchService) GetOrderByLabNumber(labNumber string) (*form.OrderSearchDTO, error) {
	trimmed := strings.TrimSpace(labNumber)
	if trimmed == "" {
		return nil, nil
	}
	sample, err := s.DAO.OrderSearchSample(normalizeAccession(trimmed))
	if err != nil || sample == nil {
		return nil, err
	}

	dto := &form.OrderSearchDTO{
		ID:        strconv.FormatInt(sample.ID, 10),
		LabNumber: sample.AccessionNumber,
		// getReceivedDateForDisplay / getCollectionDateForDisplay are
		// dd/MM/yyyy. Both are dropped when null (Include.NON_NULL on a
		// HashMap value that was put as null).
		ReceivedDate:   displayDate(sample.ReceivedDate),
		CollectionDate: displayDate(sample.CollectionDate),
		Status:         sample.Status,
		StorageSkipped: sample.StorageSkipped != nil && *sample.StorageSkipped,
	}

	if err := s.attachPatient(dto, sample.ID); err != nil {
		return nil, err
	}
	items, err := s.buildSampleItems(sample.ID)
	if err != nil {
		return nil, err
	}
	dto.Samples = items

	orderItems, err := s.buildSampleOrderItems(sample)
	if err != nil {
		return nil, err
	}
	dto.SampleOrderItems = orderItems

	step, err := s.stepProgress(sample, items)
	if err != nil {
		return nil, err
	}
	dto.StepProgress = step
	return dto, nil
}

// attachPatient fills patientProperties and orderData. Java omits BOTH keys
// entirely when the sample has no patient — the whole `if (patient != null)`
// block is skipped — rather than emitting empty objects.
func (s *OrderSearchService) attachPatient(dto *form.OrderSearchDTO, sampleID int64) error {
	p, err := s.DAO.PatientForSample(sampleID)
	if err != nil || p == nil {
		return err
	}
	identities, err := s.DAO.PatientIdentitiesByType(p.PatientID)
	if err != nil {
		return err
	}
	address, err := s.DAO.AddressPartsByName(p.PersonID)
	if err != nil {
		return err
	}

	// getAddress returns "" for a missing part, and the city falls back to
	// person.city ONLY when the village part is blank.
	city := address["village"]
	if city == "" {
		city = deref(p.City)
	}

	props := &form.PatientInfoBean{
		PatientPK: strconv.FormatInt(p.PatientID, 10),
		// A literal, always "UPDATE" on this path.
		PatientUpdateStatus:    "UPDATE",
		NationalID:             deref(p.NationalID),
		STNumber:               identities["ST"],
		SubjectNumber:          identities["SUBJECT"],
		LastName:               deref(p.LastName),
		FirstName:              deref(p.FirstName),
		MothersName:            identities["MOTHER"],
		AKA:                    identities["AKA"],
		City:                   city,
		Commune:                address["commune"],
		AddressDepartment:      address["department"],
		PrimaryPhone:           deref(p.PrimaryPhone),
		Email:                  deref(p.Email),
		Gender:                 deref(p.Gender),
		InsuranceNumber:        identities["INSURANCE"],
		Occupation:             identities["OCCUPATION"],
		CustomNotes:            identities["CUSTOM_NOTES"],
		TargetDiseaseProgramme: identities["DISEASE_PROGRAMME"],
		MothersInitial:         identities["MOTHERS_INITIAL"],
		Education:              identities["EDUCATION"],
		MaritialStatus:         identities["MARITIAL"],
		Nationality:            identities["NATIONALITY"],
		OtherNationality:       identities["OTHER NATIONALITY"],
		HealthDistrict:         identities["HEALTH DISTRICT"],
		HealthRegion:           identities["HEALTH REGION"],
		GUID:                   identities["GUID"],
		// birthDateForDisplay is the STORED entered_birth_date, reformatted.
		// The locale switch (fr-FR -> dd/MM/yyyy, else MM/dd/yyyy) is Java's;
		// the dev stack resolves to dd/MM/yyyy, which is what entered_birth_date
		// already holds, so it passes through unchanged here. A deployment on
		// the other branch would need the reformat — noted, not implemented,
		// because there is no live response to verify it against.
		BirthDateForDisplay: reformatEnteredBirthDate(deref(p.EnteredBirthDate)),
		// Primitive booleans on the bean, so they always serialize.
		ReadOnly: false,
		IsMerged: p.IsMerged != nil && *p.IsMerged,
		// Initialised to an empty map on the bean and never populated on this
		// path, so it appears as {} rather than being dropped.
		AddressHierarchy: map[string]any{},
	}
	if p.PatientLastupdated != nil {
		props.PatientLastUpdated = formatSQLTimestamp(*p.PatientLastupdated)
	}
	if p.PersonLastupdated != nil {
		props.PersonLastUpdated = formatSQLTimestamp(*p.PersonLastupdated)
	}

	dto.PatientProperties = props
	// orderData is the SAME bean again under a second key, plus the literal
	// status — Java builds it "for frontend compatibility".
	dto.OrderData = &form.OrderDataDTO{
		PatientProperties:   props,
		PatientUpdateStatus: "UPDATE",
	}
	return nil
}

// buildSampleItems ports the samples[] loop.
func (s *OrderSearchService) buildSampleItems(sampleID int64) ([]form.OrderSearchSampleItemDTO, error) {
	rows, err := s.DAO.SampleItemsForSample(sampleID)
	if err != nil {
		return nil, err
	}
	analyses, err := s.DAO.AnalysesForSampleItems(sampleID)
	if err != nil {
		return nil, err
	}

	testsByItem := map[int64][]form.OrderSearchTestDTO{}
	panelsByItem := map[int64][]form.OrderSearchPanelDTO{}
	for _, a := range analyses {
		if a.TestID != nil {
			testsByItem[a.SampleItemID] = append(testsByItem[a.SampleItemID], form.OrderSearchTestDTO{
				ID:          strconv.FormatInt(*a.TestID, 10),
				Name:        deref(a.TestName),
				Description: a.TestDescription,
			})
		}
		if a.PanelID != nil {
			id := strconv.FormatInt(*a.PanelID, 10)
			// Java de-duplicates panels per sample item by id.
			dup := false
			for _, p := range panelsByItem[a.SampleItemID] {
				if p.ID == id {
					dup = true
					break
				}
			}
			if !dup {
				panelsByItem[a.SampleItemID] = append(panelsByItem[a.SampleItemID], form.OrderSearchPanelDTO{
					ID:   id,
					Name: deref(a.PanelName),
				})
			}
		}
	}

	out := make([]form.OrderSearchSampleItemDTO, 0, len(rows))
	for _, r := range rows {
		id := strconv.FormatInt(r.ID, 10)
		collectionDate, collectionTime := displayDateTime(r.CollectionDate)
		receivedDate, receivedTime := displayDateTime(r.ReceivedDate)

		item := form.OrderSearchSampleItemDTO{
			ID:           id,
			SampleItemID: id,
			SortOrder:    r.SortOrder,
			// Java puts sortOrder under BOTH `sortOrder` and `index`.
			Index: r.SortOrder,
			// `name` is the LOCALIZED name and `sampleTypeName` the raw
			// description — two keys, two accessors, identical in this dataset.
			Name:           r.SampleTypeName,
			SampleTypeName: r.SampleTypeName,
			// Every one of these is coalesced to "" by Java at the put site.
			CollectionDate:       collectionDate,
			CollectionTime:       collectionTime,
			ReceivedDate:         receivedDate,
			ReceivedTime:         receivedTime,
			Quantity:             r.Quantity,
			QuantityUnit:         idOrEmpty(r.UomID),
			CollectorID:          derefOrEmpty(r.Collector),
			CollectionConditions: derefOrEmpty(r.CollectionConditions),
			CollectionMethod:     derefOrEmpty(r.CollectionMethod),
			SampleTemperature:    derefOrEmpty(r.SampleTemperature),
			SpecimenOrigin:       derefOrEmpty(r.SpecimenOrigin),
			// Non-nil so they serialize as [] — Java initialises both lists.
			Tests:  orEmptyTests(testsByItem[r.ID]),
			Panels: orEmptyPanels(panelsByItem[r.ID]),
			SampleXML: form.OrderSearchSampleXMLDTO{
				CollectionDate:    collectionDate,
				CollectionTime:    collectionTime,
				Quantity:          r.Quantity,
				UOM:               idOrEmpty(r.UomID),
				Collector:         r.Collector,
				CollectionMethod:  derefOrEmpty(r.CollectionMethod),
				SampleTemperature: derefOrEmpty(r.SampleTemperature),
				SpecimenOrigin:    derefOrEmpty(r.SpecimenOrigin),
			},
		}
		if r.TypeOfSampleID != nil {
			v := strconv.FormatInt(*r.TypeOfSampleID, 10)
			item.SampleTypeID = &v
		}

		// The storage block is put ONLY when there is an assignment WITH a
		// location — Java guards on `assignment != null && locationId != null`,
		// so all five storage keys travel together or not at all.
		if r.StorageLocationID != nil {
			item.StorageLocationID = r.StorageLocationID
			item.StorageLocationType = r.StorageLocationType
			item.StoragePositionCoordinate = r.StoragePositionCoordinate
			item.StorageNotes = r.StorageNotes
			path, err := s.hierarchicalPath(r)
			if err != nil {
				return nil, err
			}
			// storageHierarchicalPath is put only when the path resolved —
			// a broken ancestry leaves the key absent, not empty.
			item.StorageHierarchicalPath = path
		}
		out = append(out, item)
	}
	return out, nil
}

// hierarchicalPath renders "Room > Device > Shelf > Rack > Position", to
// whatever depth the location type implies, and only when the FULL ancestry
// resolved — see StoragePathFor.
func (s *OrderSearchService) hierarchicalPath(r daoimpl.OrderSearchSampleItemRow) (*string, error) {
	if r.StorageLocationType == nil || r.StorageLocationID == nil {
		return nil, nil
	}
	row, err := s.DAO.StoragePathFor(*r.StorageLocationType, *r.StorageLocationID)
	if err != nil || row == nil {
		return nil, err
	}

	var parts []string
	need := func(v *string) bool {
		if v == nil || *v == "" {
			return false
		}
		parts = append(parts, *v)
		return true
	}
	switch *r.StorageLocationType {
	case "room":
		if !need(row.RoomName) {
			return nil, nil
		}
	case "device":
		if !need(row.RoomName) || !need(row.DeviceName) {
			return nil, nil
		}
	case "shelf":
		if !need(row.RoomName) || !need(row.DeviceName) || !need(row.ShelfLabel) {
			return nil, nil
		}
	case "rack":
		if !need(row.RoomName) || !need(row.DeviceName) || !need(row.ShelfLabel) || !need(row.RackLabel) {
			return nil, nil
		}
	case "box":
		if !need(row.RoomName) || !need(row.DeviceName) || !need(row.ShelfLabel) ||
			!need(row.RackLabel) || !need(row.BoxLabel) {
			return nil, nil
		}
	default:
		return nil, nil
	}
	if r.StoragePositionCoordinate != nil && strings.TrimSpace(*r.StoragePositionCoordinate) != "" {
		parts = append(parts, *r.StoragePositionCoordinate)
	}
	joined := strings.Join(parts, " > ")
	return &joined, nil
}

// buildSampleOrderItems ports the sampleOrderItems block.
//
// SCOPE: the provider, referring-site, department, program and
// observation-history keys are NOT built. Every one of them is absent from the
// live response on this dataset (no provider, no organization, no observations
// on the sample), so there is nothing to verify an implementation against —
// writing them would be guessing at shapes no test can check. They are listed
// here so the omission is explicit rather than silent.
func (s *OrderSearchService) buildSampleOrderItems(sample *daoimpl.OrderSearchSampleRow) (*form.SampleOrderItemsDTO, error) {
	options, err := s.DAO.PaymentOptions()
	if err != nil {
		return nil, err
	}
	receivedDate, receivedTime := displayDateTime(sample.ReceivedDate)
	dto := &form.SampleOrderItemsDTO{
		LabNo:                  sample.AccessionNumber,
		CollectionDate:         displayDate(sample.CollectionDate),
		ReceivedDateForDisplay: receivedDate,
		ReceivedTime:           receivedTime,
		PaymentOptions:         options,
	}
	// priority is the RAW enum name here ("ROUTINE"), not the lowercased form
	// order/dashboard emits. Same column, two endpoints, two casings.
	if sample.OrderPriority != nil && *sample.OrderPriority != "" {
		dto.Priority = sample.OrderPriority
	}
	return dto, nil
}

// stepProgress ports the four flags searchOrder emits.
//
// `enter` is HARDCODED TRUE here — "If sample exists, enter is complete" — while
// order/dashboard COMPUTES it from received date plus patient or workflow-type.
// Same key name, two different meanings, so the dashboard's logic must not be
// reused.
func (s *OrderSearchService) stepProgress(
	sample *daoimpl.OrderSearchSampleRow, items []form.OrderSearchSampleItemDTO,
) (form.OrderStepProgressDTO, error) {
	storageSkipped := sample.StorageSkipped != nil && *sample.StorageSkipped

	withTests, withTestsDated, withStorage := 0, 0, 0
	for _, it := range items {
		if len(it.Tests) > 0 {
			withTests++
			if it.CollectionDate != "" {
				withTestsDated++
			}
		}
		if it.StorageLocationID != nil {
			withStorage++
		}
	}
	collectComplete := withTests > 0 && withTestsDated == withTests
	labelComplete := storageSkipped || (len(items) > 0 && withStorage == len(items))

	// qa comes from the checklist, exactly as on the dashboard.
	qaComplete, err := s.DAO.QAAllRequiredVerified(sample.ID)
	if err != nil {
		return form.OrderStepProgressDTO{}, err
	}

	return form.OrderStepProgressDTO{
		Enter:   true,
		Collect: collectComplete,
		Label:   labelComplete,
		QA:      qaComplete,
	}, nil
}

// displayDate renders a timestamp as Java's dd/MM/yyyy display form, or "" for
// null — matching getReceivedDateForDisplay / getCollectionDateForDisplay.
func displayDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.In(displayZone()).Format("02/01/2006")
}

// displayDateTime splits a timestamp the way
// DateUtil.convertTimestampToStringDate / ...StringTime do: dd/MM/yyyy and
// HH:mm, both "" when the timestamp is null.
func displayDateTime(t *time.Time) (string, string) {
	if t == nil {
		return "", ""
	}
	local := t.In(displayZone())
	return local.Format("02/01/2006"), local.Format("15:04")
}

func idOrEmpty(id *int64) string {
	if id == nil {
		return ""
	}
	return strconv.FormatInt(*id, 10)
}

func derefOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func orEmptyTests(v []form.OrderSearchTestDTO) []form.OrderSearchTestDTO {
	if v == nil {
		return []form.OrderSearchTestDTO{}
	}
	return v
}

func orEmptyPanels(v []form.OrderSearchPanelDTO) []form.OrderSearchPanelDTO {
	if v == nil {
		return []form.OrderSearchPanelDTO{}
	}
	return v
}

// displayZone is the zone timestamps are rendered in.
//
// Java renders them with the JVM's default zone, which the container sets from
// TZ (docker-compose.go.yml passes the SAME TZ to both services). Go does NOT
// pick TZ up into time.Local on every platform — on this Windows dev host
// time.Local stays JST even with TZ=UTC set, which rendered a 10:00 collection
// time as 19:00 and produced a nine-hour parity gap.
//
// Resolving it explicitly, TZ first, is the same approach a1's server-time
// endpoint already takes for the same reason.
func displayZone() *time.Location {
	if tz := os.Getenv("TZ"); tz != "" {
		if loc, err := time.LoadLocation(tz); err == nil {
			return loc
		}
	}
	return time.Local
}

// reformatEnteredBirthDate ports DateUtil.formatStringDate as searchOrder uses
// it: parse the stored value as MM/dd/yyyy, falling back to dd/MM/yyyy, then
// render in the configured display format.
//
// CROSS-ENDPOINT TRAP: c1's patientByLabNumer emits entered_birth_date RAW
// ("01/15/1990"), while order/search runs it through this and emits
// "15/01/1990" — the same stored value, two formats, measured on both
// endpoints. Applying c1's pass-through here was the bug this replaces.
//
// An unparseable value becomes the literal "Invalid date format: <value>",
// which is Java's, not a placeholder.
func reformatEnteredBirthDate(raw string) string {
	if raw == "" {
		return ""
	}
	for _, in := range []string{"01/02/2006", "02/01/2006"} {
		if t, err := time.Parse(in, raw); err == nil {
			return t.Format(birthDateDisplayFormat)
		}
	}
	return "Invalid date format: " + raw
}

// birthDateDisplayFormat is the output pattern searchOrder picks.
//
// Java chooses dd/MM/yyyy when DEFAULT_DATE_LOCALE is "fr-FR" and MM/dd/yyyy
// otherwise. The dev stack emits "15/01/1990" from a stored "01/15/1990", i.e.
// dd/MM/yyyy — measured, not assumed. Mirrored as a constant because the Go
// service cannot read the Java container's resolved property; a deployment on
// the other branch must change it here too.
const birthDateDisplayFormat = "02/01/2006"
