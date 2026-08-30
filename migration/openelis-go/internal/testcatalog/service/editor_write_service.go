package service

import (
	"errors"
	"sort"
	"strconv"
	"strings"

	"openelis-go/internal/testcatalog/daoimpl"
	"openelis-go/internal/testcatalog/form"
)

// EditorWriteService ports the section saves of
// TestCatalogEditorRestController.
//
// The controller does its own validation inline and answers 404 / 422 without
// reaching a service; those outcomes are modelled here as sentinel errors so
// the controller stays a mapping layer, which is where constitution.md Layer IV
// puts it.
type EditorWriteService struct {
	DAO *daoimpl.EditorWriteDAO
}

// ErrNotFound is the `getTestById(...) == null` / `getTypeOfSampleById(...) ==
// null` guard: 404, before anything is read or written.
var ErrNotFound = errors.New("testcatalog: not found")

// ErrUnprocessable is the 422 every inline validation in this controller
// answers with — an unknown terminology source, a blank code, a duplicate
// (source, code), an unknown panel, an empty group request.
var ErrUnprocessable = errors.New("testcatalog: unprocessable")

// ---------------------------------------------------------------- storage

// GetStorage ports getStorage.
//
// A test with no handling row is NOT a 404: the section renders blank, so the
// response is a document carrying the test id and the four flags explicitly
// false. Every other field is absent.
func (s *EditorWriteService) GetStorage(testID string) (*form.StorageDTO, error) {
	ok, err := s.DAO.TestExists(testID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotFound
	}
	stored, err := s.DAO.GetStorage(testID)
	if err != nil {
		return nil, err
	}
	return storageDTO(testID, stored), nil
}

// SaveStorage ports saveStorage.
func (s *EditorWriteService) SaveStorage(testID string, body form.StorageDTO, sysUserID int64) (*form.StorageDTO, error) {
	ok, err := s.DAO.TestExists(testID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotFound
	}
	saved, err := s.DAO.SaveStorage(testID, toHandling(body), sysUserID)
	if err != nil {
		return nil, err
	}
	return storageDTO(testID, saved), nil
}

// SaveGroupStorage ports saveGroupStorage.
//
// The guard is on the REQUEST, not on the tests: a missing storage document or
// an empty id list is 422. A test id that names nothing is then skipped in
// silence, so a request naming one real and one imaginary test is a 200 that
// wrote once. The response has no body at all.
func (s *EditorWriteService) SaveGroupStorage(body form.GroupStorageUpdate, sysUserID int64) error {
	if len(body.TestIDs) == 0 || body.Storage == nil {
		return ErrUnprocessable
	}
	desired := toHandling(*body.Storage)
	for _, testID := range body.TestIDs {
		ok, err := s.DAO.TestExists(testID)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		if _, err := s.DAO.SaveStorage(testID, desired, sysUserID); err != nil {
			return err
		}
	}
	return nil
}

// toHandling ports the controller's own toHandling: a BLANK string becomes
// NULL, and a missing boolean becomes false rather than staying unset —
// `Boolean.TRUE.equals(x)` is false for null.
func toHandling(body form.StorageDTO) daoimpl.StorageRow {
	return daoimpl.StorageRow{
		StorageCondition:       nullIfBlank(body.StorageCondition),
		StorageConditionCustom: nullIfBlank(body.StorageConditionCustom),
		StorageDuration:        body.StorageDuration,
		StorageDurationUnit:    nullIfBlank(body.StorageDurationUnit),
		StabilityNotes:         nullIfBlank(body.StabilityNotes),
		ProtectFromLight:       isTrue(body.ProtectFromLight),
		DoNotFreeze:            isTrue(body.DoNotFreeze),
		DoNotRefrigerate:       isTrue(body.DoNotRefrigerate),
		DisposalMethod:         nullIfBlank(body.DisposalMethod),
		DisposalTimeframe:      body.DisposalTimeframe,
		DisposalUnit:           nullIfBlank(body.DisposalUnit),
		SpecialInstructions:    nullIfBlank(body.SpecialInstructions),
		OverrideRestricted:     isTrue(body.OverrideRestricted),
	}
}

func storageDTO(testID string, stored daoimpl.StoredStorage) *form.StorageDTO {
	dto := &form.StorageDTO{TestID: testID}
	if !stored.Found {
		f := false
		dto.ProtectFromLight, dto.DoNotFreeze = &f, &f
		dnr, or := false, false
		dto.DoNotRefrigerate, dto.OverrideRestricted = &dnr, &or
		return dto
	}
	r := stored.Row
	dto.StorageCondition = r.StorageCondition
	dto.StorageConditionCustom = r.StorageConditionCustom
	dto.StorageDuration = r.StorageDuration
	dto.StorageDurationUnit = r.StorageDurationUnit
	dto.StabilityNotes = r.StabilityNotes
	pfl, dnf, dnr, or := r.ProtectFromLight, r.DoNotFreeze, r.DoNotRefrigerate, r.OverrideRestricted
	dto.ProtectFromLight, dto.DoNotFreeze = &pfl, &dnf
	dto.DoNotRefrigerate, dto.OverrideRestricted = &dnr, &or
	dto.DisposalMethod = r.DisposalMethod
	dto.DisposalTimeframe = r.DisposalTimeframe
	dto.DisposalUnit = r.DisposalUnit
	dto.SpecialInstructions = r.SpecialInstructions
	return dto
}

// ------------------------------------------------------------ terminology

// termSources and termRelationships are the controller's allow-lists. A source
// outside the first, or a relationship outside the second, is 422 — the DB
// constrains neither.
var termSources = map[string]bool{"LOINC": true, "SNOMED": true, "CIEL": true, "OCL": true}

var termRelationships = map[string]bool{"SAME_AS": true, "BROADER_THAN": true, "NARROWER_THAN": true}

// GetTerminology ports getTerminology.
func (s *EditorWriteService) GetTerminology(testID string) (*form.TerminologyResponse, error) {
	ok, err := s.DAO.TestExists(testID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotFound
	}
	return s.terminologyResponse(testID)
}

// SaveTerminology ports saveTerminology.
//
// Validation runs over the WHOLE list before any write, so a request whose last
// entry is bad writes nothing at all.
func (s *EditorWriteService) SaveTerminology(testID string, body form.TerminologyResponse, sysUserID int64) (*form.TerminologyResponse, error) {
	ok, err := s.DAO.TestExists(testID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotFound
	}

	seen := map[string]bool{}
	desired := make([]daoimpl.DesiredMapping, 0, len(body.Mappings))
	for _, m := range body.Mappings {
		if strings.TrimSpace(m.Source) == "" || !termSources[m.Source] || strings.TrimSpace(m.Code) == "" {
			return nil, ErrUnprocessable
		}
		relationship := ""
		if m.Relationship != nil {
			relationship = *m.Relationship
		}
		if strings.TrimSpace(relationship) != "" && !termRelationships[relationship] {
			return nil, ErrUnprocessable
		}
		key := m.Source + " " + m.Code
		if seen[key] {
			return nil, ErrUnprocessable
		}
		seen[key] = true

		want := daoimpl.DesiredMapping{Source: m.Source, Code: m.Code}
		// A BLANK relationship is stored as NULL, not as "".
		if strings.TrimSpace(relationship) != "" {
			r := relationship
			want.Relationship = &r
		}
		desired = append(desired, want)
	}

	if err := s.DAO.SaveTerminology(testID, desired, sysUserID); err != nil {
		return nil, err
	}
	return s.terminologyResponse(testID)
}

func (s *EditorWriteService) terminologyResponse(testID string) (*form.TerminologyResponse, error) {
	rows, err := s.DAO.ActiveTerminology(testID)
	if err != nil {
		return nil, err
	}
	resp := &form.TerminologyResponse{TestID: testID, Mappings: []form.MappingDTO{}}
	for _, r := range rows {
		resp.Mappings = append(resp.Mappings, form.MappingDTO{
			ID: r.ID, Source: r.Source, Code: r.Code, Relationship: r.Relationship,
		})
	}
	return resp, nil
}

// ------------------------------------------------------- sample-type order

// GetTestOrder ports getTestOrder.
func (s *EditorWriteService) GetTestOrder(sampleTypeID string) (*form.DisplayOrderResponse, error) {
	ok, err := s.DAO.SampleTypeExists(sampleTypeID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotFound
	}
	return s.testOrderResponse(sampleTypeID)
}

// SaveTestOrder ports saveTestOrder.
//
// An entry with a blank test id or a null order is dropped from the map, so it
// is not an error and not a write — it just does not participate.
func (s *EditorWriteService) SaveTestOrder(sampleTypeID string, body form.DisplayOrderUpdate) (*form.DisplayOrderResponse, error) {
	ok, err := s.DAO.SampleTypeExists(sampleTypeID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotFound
	}
	orderByTestID := map[string]int{}
	for _, item := range body.Items {
		if strings.TrimSpace(item.TestID) == "" || item.DisplayOrder == nil {
			continue
		}
		orderByTestID[item.TestID] = *item.DisplayOrder
	}
	if err := s.DAO.SaveTestOrder(sampleTypeID, orderByTestID); err != nil {
		return nil, err
	}
	return s.testOrderResponse(sampleTypeID)
}

func (s *EditorWriteService) testOrderResponse(sampleTypeID string) (*form.DisplayOrderResponse, error) {
	rows, err := s.DAO.TestOrder(sampleTypeID)
	if err != nil {
		return nil, err
	}
	resp := &form.DisplayOrderResponse{SampleTypeID: sampleTypeID, Tests: []form.TestOrderRowDTO{}}
	for _, r := range rows {
		resp.Tests = append(resp.Tests, form.TestOrderRowDTO{
			TestID: r.TestID, TestName: r.TestName, DisplayOrder: r.DisplayOrder,
		})
	}
	// displayOrder ascending with nulls LAST (a null sorts as MAX_VALUE), then
	// the test name case-insensitively.
	sort.SliceStable(resp.Tests, func(i, j int) bool {
		a, b := orderOrMax(resp.Tests[i].DisplayOrder), orderOrMax(resp.Tests[j].DisplayOrder)
		if a != b {
			return a < b
		}
		return strings.ToLower(derefString(resp.Tests[i].TestName)) <
			strings.ToLower(derefString(resp.Tests[j].TestName))
	})
	return resp, nil
}

// ------------------------------------------------------------------ panels

// GetTestPanels ports getTestPanels.
func (s *EditorWriteService) GetTestPanels(testID string) (*form.TestPanelsResponse, error) {
	ok, err := s.DAO.TestExists(testID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotFound
	}
	return s.testPanelsResponse(testID)
}

// SaveTestPanels ports saveTestPanels.
//
// The position fallback is the trap. `fallback` starts at 1 and increments once
// per ITEM — including items with a blank panel id — so a membership sent with
// a null position takes its 1-based INDEX in the request, not its rank among
// the memberships that were kept.
func (s *EditorWriteService) SaveTestPanels(testID string, body form.PanelMembershipUpdate) (*form.TestPanelsResponse, error) {
	ok, err := s.DAO.TestExists(testID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotFound
	}

	positionByPanelID := map[string]*int{}
	fallback := 1
	for _, item := range body.Memberships {
		if strings.TrimSpace(item.PanelID) != "" {
			exists, err := s.DAO.PanelExists(item.PanelID)
			if err != nil {
				return nil, err
			}
			if !exists {
				// Rejected up front rather than dropped silently by the service.
				return nil, ErrUnprocessable
			}
			position := item.Position
			if position == nil {
				f := fallback
				position = &f
			}
			positionByPanelID[item.PanelID] = position
		}
		fallback++
	}

	if err := s.DAO.SaveTestPanels(testID, positionByPanelID); err != nil {
		return nil, err
	}
	return s.testPanelsResponse(testID)
}

func (s *EditorWriteService) testPanelsResponse(testID string) (*form.TestPanelsResponse, error) {
	rows, err := s.DAO.TestPanels(testID)
	if err != nil {
		return nil, err
	}
	resp := &form.TestPanelsResponse{TestID: testID, Memberships: []form.PanelMembershipDTO{}}
	for _, r := range rows {
		resp.Memberships = append(resp.Memberships, form.PanelMembershipDTO{
			PanelID: r.PanelID, PanelName: r.PanelName, Position: parseIntOrNil(r.SortOrder),
		})
	}
	// By panel name, case-insensitively. Nothing breaks a tie.
	sort.SliceStable(resp.Memberships, func(i, j int) bool {
		return strings.ToLower(derefString(resp.Memberships[i].PanelName)) <
			strings.ToLower(derefString(resp.Memberships[j].PanelName))
	})
	return resp, nil
}

// GetPanelTestOrder ports getPanelTestOrder — read-only, the preview beside the
// position editor.
func (s *EditorWriteService) GetPanelTestOrder(panelID string) (*form.PanelTestOrderResponse, error) {
	ok, err := s.DAO.PanelExists(panelID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotFound
	}
	rows, err := s.DAO.PanelTests(panelID)
	if err != nil {
		return nil, err
	}
	resp := &form.PanelTestOrderResponse{PanelID: panelID, Tests: []form.PanelTestRowDTO{}}
	for _, r := range rows {
		resp.Tests = append(resp.Tests, form.PanelTestRowDTO{
			TestID: r.TestID, TestName: r.TestName, Position: r.DisplayOrder,
		})
	}
	sort.SliceStable(resp.Tests, func(i, j int) bool {
		a, b := orderOrMax(resp.Tests[i].Position), orderOrMax(resp.Tests[j].Position)
		if a != b {
			return a < b
		}
		return strings.ToLower(derefString(resp.Tests[i].TestName)) <
			strings.ToLower(derefString(resp.Tests[j].TestName))
	})
	return resp, nil
}

// CreatePanel ports createPanel.
//
// A blank name is 422. ANY other name is a 500: the insert leaves
// name_localization_id null and the column is NOT NULL. See the DAO — the
// failure is Java's and it is reproduced, not repaired.
func (s *EditorWriteService) CreatePanel(body form.CreatePanelRequest) (*form.IdNameDTO, error) {
	if strings.TrimSpace(body.Name) == "" {
		return nil, ErrUnprocessable
	}
	id, err := s.DAO.CreatePanel(body.Name)
	if err != nil {
		return nil, err
	}
	return &form.IdNameDTO{ID: id, Name: strings.TrimSpace(body.Name)}, nil
}

// ----------------------------------------------------------------- helpers

func nullIfBlank(s *string) *string {
	if s == nil || strings.TrimSpace(*s) == "" {
		return nil
	}
	return s
}

func isTrue(b *bool) bool { return b != nil && *b }

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// orderOrMax is Java's `x != null ? x : Integer.MAX_VALUE`.
func orderOrMax(v *int) int {
	if v == nil {
		return 2147483647
	}
	return *v
}

// parseIntOrNil is the controller's parseIntOrNull: a blank or non-numeric
// sort order reads as absent rather than as zero.
func parseIntOrNil(s *string) *int {
	if s == nil || strings.TrimSpace(*s) == "" {
		return nil
	}
	n, err := strconv.Atoi(strings.TrimSpace(*s))
	if err != nil {
		return nil
	}
	return &n
}
