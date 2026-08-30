package service

import (
	"encoding/json"
	"sort"

	commondaoimpl "openelis-go/internal/common/daoimpl"
	"openelis-go/internal/common/util"
	"openelis-go/internal/testconfiguration/daoimpl"
	"openelis-go/internal/testconfiguration/form"
)

// ActivationService ports TestActivationRestController and
// TestOrderabilityRestController.
//
// Both read the same double-encoded jsonChangeList the *Order screens use, and
// both answer with a list built per SAMPLE TYPE rather than a flat one.
type ActivationService struct {
	Lists *commondaoimpl.DisplayListDAOImpl
	DAO   *daoimpl.ActivationDAO
}

// idsFor is getIdsForActions: the same double encoding, but the inner array
// holds OBJECTS with an id rather than plain values.
//
// requireStringID is TestOrderability. Its copy of the method does
// `list.add((String) ((JSONObject) actionArray.get(i)).get("id"))` — a cast of
// the id VALUE to String, so a numeric id is a ClassCastException and a 500.
// TestActivation.s copy does not cast, and takes a number happily. Two methods
// with one name, and the difference is only visible from outside as a 500.
func idsFor(changeList, key string, requireStringID bool) ([]string, error) {
	var root map[string]json.RawMessage
	if err := json.Unmarshal([]byte(changeList), &root); err != nil {
		return nil, nil
	}
	raw, ok := root[key]
	if !ok {
		// `String action = (String) root.get(key)` yields null for a missing key
		// and `parser.parse(null)` throws NullPointerException — not the
		// ParseException the catch is written for. An omitted key is a 500.
		return nil, ErrChangeListShape
	}
	var inner string
	if err := json.Unmarshal(raw, &inner); err != nil {
		return nil, ErrChangeListShape
	}
	var entries []map[string]json.RawMessage
	if err := json.Unmarshal([]byte(inner), &entries); err != nil {
		return nil, nil
	}
	ids := make([]string, 0, len(entries))
	for _, e := range entries {
		idRaw, has := e["id"]
		if !has {
			continue
		}
		var asString string
		if err := json.Unmarshal(idRaw, &asString); err == nil {
			ids = append(ids, asString)
			continue
		}
		if requireStringID {
			return nil, ErrChangeListShape
		}
		var asNumber json.Number
		if err := json.Unmarshal(idRaw, &asNumber); err != nil {
			return nil, ErrChangeListShape
		}
		ids = append(ids, asNumber.String())
	}
	return ids, nil
}

// TestActivationForm ports showTestActivation.
func (s *ActivationService) TestActivationForm() (*form.ActivationForm, error) {
	active, inactive, err := s.testLists()
	if err != nil {
		return nil, err
	}
	f := form.NewActivationForm("testActivationForm")
	f.ActiveTestList, f.InactiveTestList = &active, &inactive
	return &f, nil
}

// testLists ports createTestList(true) and createTestList(false).
//
// The ACTIVE list walks SAMPLE_TYPE_ACTIVE in its display order; the INACTIVE
// one walks SAMPLE_TYPE_INACTIVE re-sorted ALPHABETICALLY — "if not active we
// use alphabetical ordering, the default is display order". Within each sample
// type the tests are sorted by numeric sort order, split into active and
// inactive, and the inactive half is re-sorted by name afterwards.
func (s *ActivationService) testLists() ([]form.TestActivationBean, []form.TestActivationBean, error) {
	activeTypes, err := s.Lists.ActiveHumanSampleTypes()
	if err != nil {
		return nil, nil, err
	}
	inactiveTypes, err := s.Lists.InactiveHumanSampleTypes()
	if err != nil {
		return nil, nil, err
	}
	sort.SliceStable(inactiveTypes, func(i, j int) bool {
		return inactiveTypes[i].Value < inactiveTypes[j].Value
	})

	tests, err := s.DAO.TestsBySampleType()
	if err != nil {
		return nil, nil, err
	}
	bySampleType := map[string][]daoimpl.TestRow{}
	for _, t := range tests {
		bySampleType[t.SampleTypeID] = append(bySampleType[t.SampleTypeID], t)
	}

	build := func(types []util.IdValuePair) []form.TestActivationBean {
		out := make([]form.TestActivationBean, 0, len(types))
		for _, st := range types {
			sorted := daoimpl.SortTestsForDisplay(bySampleType[st.Id])
			activeTests := []util.IdValuePair{}
			inactiveTests := []util.IdValuePair{}
			for _, t := range sorted {
				pair := util.NewIdValuePair(t.ID, t.Name)
				if t.IsActive == "Y" {
					activeTests = append(activeTests, pair)
				} else {
					inactiveTests = append(inactiveTests, pair)
				}
			}
			sort.SliceStable(inactiveTests, func(i, j int) bool {
				return inactiveTests[i].Value < inactiveTests[j].Value
			})
			out = append(out, form.TestActivationBean{
				SampleType:    st,
				ActiveTests:   activeTests,
				InactiveTests: inactiveTests,
			})
		}
		return out
	}
	return build(activeTypes), build(inactiveTypes), nil
}

// ApplyTestActivation ports postTestActivation.
//
// Four keys, two shapes: activateSample and activateTest carry {id, sortOrder};
// deactivateSample and deactivateTest carry plain ids. An activation multiplies
// the submitted order by TEN before storing it —
// setSortOrder(String.valueOf(set.sortOrder * 10)) — so the stored value is
// never the one the client sent.
func (s *ActivationService) ApplyTestActivation(post form.ActivationPost, sysUserID int64) (*form.ActivationForm, error) {
	f := form.NewActivationForm("testActivationForm")
	if post.JSONChangeList != nil {
		f.JSONChangeList = *post.JSONChangeList
	}

	activateSamples, err := parseChangeList(f.JSONChangeList, "activateSample")
	if err != nil {
		return nil, err
	}
	deactivateSamples, err := idsFor(f.JSONChangeList, "deactivateSample", false)
	if err != nil {
		return nil, err
	}
	activateTests, err := parseChangeList(f.JSONChangeList, "activateTest")
	if err != nil {
		return nil, err
	}
	deactivateTests, err := idsFor(f.JSONChangeList, "deactivateTest", false)
	if err != nil {
		return nil, err
	}

	// TestActivationFormValidator runs over the whole list first: every
	// activateTest entry needs a valid id, a sort order of at most three digits,
	// and an `activated` field the handler never reads. A single failure refuses
	// the request — 200, nothing written.
	for _, set := range activateTests {
		if !idFormat.MatchString(set.ID.String()) ||
			!sortOrderFormat.MatchString(set.SortOrder.String()) ||
			set.Activated == nil {
			active, inactive, listErr := s.testLists()
			if listErr != nil {
				return nil, listErr
			}
			f.ActiveTestList, f.InactiveTestList = &active, &inactive
			return &f, nil
		}
	}

	testChanges := []daoimpl.TestChange{}
	for _, id := range deactivateTests {
		testChanges = append(testChanges, daoimpl.TestChange{ID: id, Active: false})
	}
	for _, set := range activateTests {
		n, convErr := set.SortOrder.Int64()
		if convErr != nil {
			continue
		}
		testChanges = append(testChanges, daoimpl.TestChange{
			ID: set.ID.String(), Active: true,
			SortOrder: int(n) * 10, SortOrderSet: true,
		})
	}

	typeChanges := []daoimpl.SampleTypeChange{}
	for _, id := range deactivateSamples {
		typeChanges = append(typeChanges, daoimpl.SampleTypeChange{ID: id, Active: false})
	}
	for _, set := range activateSamples {
		n, convErr := set.SortOrder.Int64()
		if convErr != nil {
			continue
		}
		typeChanges = append(typeChanges, daoimpl.SampleTypeChange{
			ID: set.ID.String(), Active: true,
			SortOrder: int(n) * 10, SortOrderSet: true,
		})
	}

	if len(testChanges) > 0 || len(typeChanges) > 0 {
		if err := s.DAO.Apply(testChanges, typeChanges, sysUserID); err != nil {
			// LIMSRuntimeException is caught and logged at DEBUG.
			_ = err
		}
	}

	// The POST answers the REBUILT lists, not the bound form — createTestList
	// runs again after the write, so the response shows the new state.
	active, inactive, err := s.testLists()
	if err != nil {
		return nil, err
	}
	f.ActiveTestList, f.InactiveTestList = &active, &inactive
	return &f, nil
}

// TestOrderabilityForm ports showTestOrderability — ORDERABLE_TESTS, built the
// same way as the activation lists but split on `orderable` rather than
// `is_active`.
func (s *ActivationService) TestOrderabilityForm() (*form.ActivationForm, error) {
	list, err := s.orderableLists()
	if err != nil {
		return nil, err
	}
	f := form.NewActivationForm("testOrderabilityForm")
	f.OrderableTestList = &list
	return &f, nil
}

func (s *ActivationService) orderableLists() ([]form.TestActivationBean, error) {
	types, err := s.Lists.ActiveHumanSampleTypes()
	if err != nil {
		return nil, err
	}
	tests, err := s.DAO.TestsBySampleType()
	if err != nil {
		return nil, err
	}
	bySampleType := map[string][]daoimpl.TestRow{}
	for _, t := range tests {
		bySampleType[t.SampleTypeID] = append(bySampleType[t.SampleTypeID], t)
	}
	out := make([]form.TestActivationBean, 0, len(types))
	for _, st := range types {
		sorted := daoimpl.SortTestsForDisplay(bySampleType[st.Id])
		orderable := []util.IdValuePair{}
		unorderable := []util.IdValuePair{}
		for _, t := range sorted {
			pair := util.NewIdValuePair(t.ID, t.Name)
			if t.Orderable {
				orderable = append(orderable, pair)
			} else {
				unorderable = append(unorderable, pair)
			}
		}
		sort.SliceStable(unorderable, func(i, j int) bool {
			return unorderable[i].Value < unorderable[j].Value
		})
		out = append(out, form.TestActivationBean{
			SampleType:    st,
			ActiveTests:   orderable,
			InactiveTests: unorderable,
		})
	}
	return out, nil
}

// ApplyTestOrderability ports postTestOrderability: activateTest makes a test
// orderable, deactivateTest makes it not. Two keys, plain ids, no sort order.
func (s *ActivationService) ApplyTestOrderability(post form.ActivationPost, sysUserID int64) (*form.ActivationForm, error) {
	f := form.NewActivationForm("testOrderabilityForm")
	if post.JSONChangeList != nil {
		f.JSONChangeList = *post.JSONChangeList
	}

	orderable, err := idsFor(f.JSONChangeList, "activateTest", true)
	if err != nil {
		return nil, err
	}
	unorderable, err := idsFor(f.JSONChangeList, "deactivateTest", true)
	if err != nil {
		return nil, err
	}

	// Java builds the unorderable list FIRST and appends the orderable one, so
	// an id in both lists ends up orderable.
	if err := s.DAO.SetOrderable(unorderable, false, sysUserID); err != nil {
		_ = err
	}
	if err := s.DAO.SetOrderable(orderable, true, sysUserID); err != nil {
		_ = err
	}

	list, err := s.orderableLists()
	if err != nil {
		return nil, err
	}
	f.OrderableTestList = &list
	return &f, nil
}
