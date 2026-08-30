package service

import (
	"sort"
	"strings"

	commondaoimpl "openelis-go/internal/common/daoimpl"
	"openelis-go/internal/common/util"
	"openelis-go/internal/testconfiguration/daoimpl"
	"openelis-go/internal/testconfiguration/form"
)

// AssignService ports SampleTypeTestAssignRestController,
// TestSectionTestAssignRestController and PanelTestAssignRestController.
type AssignService struct {
	Lists      *commondaoimpl.DisplayListDAOImpl
	Activation *daoimpl.ActivationDAO
	DAO        *daoimpl.AssignDAO
}

// blankPair is getListWithLeadingBlank's extra first entry: id "0", value "".
var blankPair = util.NewIdValuePair("0", "")

// SampleTypeForm ports SampleTypeTestAssign's setupDisplayItems.
//
// sampleTypeList is TWO lists concatenated — SAMPLE_TYPE with a leading blank,
// then SAMPLE_TYPE_INACTIVE — so an inactive type appears TWICE when it is
// already in the first list. The comment above it says only that the cached
// list must not be appended to in place.
func (s *AssignService) SampleTypeForm() (*form.AssignForm, error) {
	all, err := s.Lists.AllSampleTypes()
	if err != nil {
		return nil, err
	}
	inactive, err := s.Lists.InactiveHumanSampleTypes()
	if err != nil {
		return nil, err
	}
	withBlank := append([]util.IdValuePair{blankPair}, all...)
	joined := append(append([]util.IdValuePair{}, withBlank...), inactive...)

	tests, err := s.Activation.TestsBySampleType()
	if err != nil {
		return nil, err
	}
	byType := map[string][]daoimpl.TestRow{}
	for _, t := range tests {
		byType[t.SampleTypeID] = append(byType[t.SampleTypeID], t)
	}

	// The map KEY is the IdValuePair.toString() — Jackson renders a non-String
	// map key that way — and only ACTIVE tests are listed under it.
	rendered := map[string][]util.IdValuePair{}
	for _, st := range withBlank {
		key := "id=" + st.Id + ", value=" + st.Value
		entries := []util.IdValuePair{}
		for _, t := range byType[st.Id] {
			if t.IsActive != "Y" {
				continue
			}
			entries = append(entries, util.NewIdValuePair(t.ID, t.AugmentedName))
		}
		rendered[key] = entries
	}

	f := form.NewAssignForm("sampleTypeTestAssignForm")
	f.SampleTypeList = &joined
	f.SampleTypeTestList = &rendered
	empty := ""
	f.SampleTypeID, f.DeactivateSampleTypeID = &empty, &empty
	return &f, nil
}

// TestSectionForm ports TestSectionTestAssign's setupDisplayItems — the same
// shape with test sections in place of sample types.
func (s *AssignService) TestSectionForm() (*form.AssignForm, error) {
	active, err := s.Lists.ActiveTestSections()
	if err != nil {
		return nil, err
	}
	inactive, err := s.Lists.InactiveTestSections()
	if err != nil {
		return nil, err
	}
	withBlank := append([]util.IdValuePair{blankPair}, active...)
	joined := append(append([]util.IdValuePair{}, withBlank...), inactive...)

	bySection, err := s.Activation.TestsByTestSection()
	if err != nil {
		return nil, err
	}
	rendered := map[string][]util.IdValuePair{}
	for _, ts := range withBlank {
		key := "id=" + ts.Id + ", value=" + ts.Value
		entries := []util.IdValuePair{}
		for _, t := range bySection[ts.Id] {
			if t.IsActive != "Y" {
				continue
			}
			entries = append(entries, util.NewIdValuePair(t.ID, t.AugmentedName))
		}
		rendered[key] = entries
	}

	f := form.NewAssignForm("testSectionTestAssignForm")
	f.TestSectionList = &joined
	f.SectionTestList = &rendered
	empty := ""
	f.TestSectionID, f.DeactivateTestSectionID = &empty, &empty
	return &f, nil
}

// PanelForm ports PanelTestAssign's setupDisplayItems.
//
// selectedPanel is ALWAYS present, even on the blank form, where it is
// {"tests":[],"availableTests":[]}. setupDisplayItems only fills it when a
// panel id is set, but the bean initialises the object either way.
func (s *AssignService) PanelForm(panelID string) (*form.AssignForm, error) {
	panels, err := s.Lists.Panels()
	if err != nil {
		return nil, err
	}
	f := form.NewAssignForm("panelTestAssignForm")
	f.PanelList = &panels
	empty := ""
	f.PanelID, f.DeactivatePanelID = &empty, &empty

	f.SelectedPanel = &form.SelectedPanelDTO{
		Tests: []util.IdValuePair{}, AvailableTests: []util.IdValuePair{},
	}
	if strings.TrimSpace(panelID) == "" {
		return &f, nil
	}

	items, err := s.DAO.PanelItems(panelID)
	if err != nil {
		return nil, err
	}
	all, err := s.Activation.AllTests()
	if err != nil {
		return nil, err
	}
	inPanel := map[string]bool{}
	tests := []util.IdValuePair{}
	for _, it := range items {
		inPanel[it.TestID] = true
	}
	available := []util.IdValuePair{}
	for _, t := range all {
		if t.IsActive != "Y" {
			continue
		}
		pair := util.NewIdValuePair(t.ID, t.Name)
		if inPanel[t.ID] {
			tests = append(tests, pair)
		} else {
			available = append(available, pair)
		}
	}
	sort.SliceStable(available, func(i, j int) bool {
		return available[i].Value < available[j].Value
	})
	f.SelectedPanel = &form.SelectedPanelDTO{Tests: tests, AvailableTests: available}
	return &f, nil
}

// AssignSampleType ports postSampleTypeTestAssign.
//
// Assigning a test to the sample type it is being moved AWAY from is a no-op:
// `if (sampleTypeId.equals(deactivateSampleTypeId)) return form;` — the guard
// comes before any write, so nothing happens and the caller is told 200.
func (s *AssignService) AssignSampleType(post form.AssignPost, sysUserID int64) (*form.AssignForm, error) {
	f := form.NewAssignForm("sampleTypeTestAssignForm")
	f.TestID, f.SampleTypeID, f.DeactivateSampleTypeID =
		post.TestID, post.SampleTypeID, post.DeactivateSampleTypeID

	testID, sampleTypeID := derefOr(post.TestID), derefOr(post.SampleTypeID)
	deactivate := derefOr(post.DeactivateSampleTypeID)
	if sampleTypeID == deactivate {
		return &f, nil
	}
	if testID == "" || sampleTypeID == "" {
		return &f, nil
	}
	if err := s.DAO.AssignSampleType(testID, sampleTypeID, deactivate, sysUserID); err != nil {
		// HibernateException is caught and logged.
		return &f, nil
	}
	return &f, nil
}

// AssignTestSection ports postTestSectionTestAssign, with the same
// equal-ids guard.
func (s *AssignService) AssignTestSection(post form.AssignPost, sysUserID int64) (*form.AssignForm, error) {
	f := form.NewAssignForm("testSectionTestAssignForm")
	f.TestID, f.TestSectionID, f.DeactivateTestSectionID =
		post.TestID, post.TestSectionID, post.DeactivateTestSectionID

	testID, sectionID := derefOr(post.TestID), derefOr(post.TestSectionID)
	deactivate := derefOr(post.DeactivateTestSectionID)
	if sectionID == deactivate {
		return &f, nil
	}
	if testID == "" || sectionID == "" {
		return &f, nil
	}
	if err := s.DAO.AssignTestSection(testID, sectionID, deactivate, sysUserID); err != nil {
		return &f, nil
	}
	return &f, nil
}

// AssignPanelTests ports postPanelTestAssign: the panel's membership is
// REPLACED by currentTests, so an empty list empties the panel.
func (s *AssignService) AssignPanelTests(post form.AssignPost, sysUserID int64) (*form.AssignForm, error) {
	f := form.NewAssignForm("panelTestAssignForm")
	f.PanelID, f.DeactivatePanelID = post.PanelID, post.DeactivatePanelID

	panelID := derefOr(post.PanelID)
	if strings.TrimSpace(panelID) == "" {
		// `if (!isBlankOrNull(panelId))` — a blank panel id skips the whole
		// block and writes nothing.
		return &f, nil
	}
	if err := s.DAO.AssignPanelTests(panelID, post.CurrentTests, sysUserID); err != nil {
		return &f, nil
	}
	return &f, nil
}
