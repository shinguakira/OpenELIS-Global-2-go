// Package service ports the four WorkPlanBy* controllers' assembly logic
// (constitution.md Layer III). Folder layout mirrors the Java source.
package service

import (
	"errors"
	"sort"
	"strconv"

	"openelis-go/internal/common/util"
	"openelis-go/internal/workplan/daoimpl"
	"openelis-go/internal/workplan/form"
)

// WorkplanService builds WorkplanForm for all four routes.
type WorkplanService struct {
	DAO *daoimpl.WorkplanDAOImpl
}

// emptyString is the value getPatientName and getSubjectNumber return on this
// deployment: the first is gated on configurationName == "Haiti LNSP" and the
// second on SUBJECT_ON_WORKPLAN == "true", and neither holds here. They are
// separate fields with separate gates, so they are not folded together.
const emptyString = ""

// ByTest ports WorkplanByTestRestController.
//
// It sets patientName and never sets testName — the mirror image of ByPanel and
// ByTestSection. Four controllers, four different subsets of the same bean.
func (s *WorkplanService) ByTest(testID string) (*form.WorkplanForm, error) {
	if testID == "" || testID == "0" {
		f := form.NewWorkplanForm()
		return &f, nil
	}
	rows, err := s.DAO.ByTest(testID)
	if err != nil {
		return nil, err
	}
	return assemble(rows, options{
		// setTestId(testType) — the REQUESTED id, echoed onto every row, not
		// the analysis's own test id. Identical here because the query filters
		// on it, but it is what the controller writes.
		forcedTestID:   testID,
		setPatientName: true,
	}), nil
}

// ByPanel ports WorkplanByPanelRestController: expand the panel to its member
// tests, run the ByTest query per member, concatenate.
func (s *WorkplanService) ByPanel(panelID string) (*form.WorkplanForm, error) {
	if panelID == "" || panelID == "0" {
		f := form.NewWorkplanForm()
		return &f, nil
	}
	tests, err := s.DAO.PanelMemberTests(panelID)
	if err != nil {
		return nil, err
	}
	all := []daoimpl.WorkplanAnalysisRow{}
	for _, t := range tests {
		rows, err := s.DAO.ByTest(t)
		if err != nil {
			return nil, err
		}
		all = append(all, rows...)
	}
	// ByPanel is the ONLY one of the four that sorts BEFORE it numbers, and it
	// sorts on the accession ALONE with no tie-break. The other three number in
	// query order and sort afterwards, so a row's grouping number there records
	// where its accession sat in the DATABASE's ordering. Here it records where
	// the accession sits in the SORTED list, which is why rows sharing an
	// accession share a number — they have been made adjacent first.
	//
	// Concatenating per-member-test results makes accessions repeat, so numbering
	// before the sort would restart the counter on every repeat.
	return assembleSortedFirst(all, options{setTestName: true}), nil
}

// ByTestSection ports WorkplanByTestSectionRestController.
func (s *WorkplanService) ByTestSection(sectionID string) (*form.WorkplanForm, error) {
	if sectionID == "" || sectionID == "0" {
		f := form.NewWorkplanForm()
		return &f, nil
	}
	// JAVA DEFECT, reproduced: getTestDisplayName dereferences
	// sampleItem.getTypeOfSampleId() with no null check, so a single analysis on
	// a type-less sample item makes the whole request throw. Checked up front
	// because the port has to fail on exactly the same inputs, not tolerate a
	// null Java does not tolerate.
	poisoned, err := s.DAO.SectionHasTypelessItem(sectionID)
	if err != nil {
		return nil, err
	}
	if poisoned {
		return nil, ErrTypelessSampleItem
	}
	rows, err := s.DAO.ByTestSectionAugmented(sectionID)
	if err != nil {
		return nil, err
	}
	return assemble(rows, options{setTestName: true}), nil
}

// ByPriority ports WorkplanByPriorityRestController, the only one of the four
// that sets BOTH patientName and testName.
func (s *WorkplanService) ByPriority(priority string) (*form.WorkplanForm, error) {
	if priority == "" {
		f := form.NewWorkplanForm()
		return &f, nil
	}
	rows, err := s.DAO.ByPriority(priority)
	if err != nil {
		return nil, err
	}
	return assemble(rows, options{setPatientName: true, setTestName: true}), nil
}

type options struct {
	forcedTestID   string
	setPatientName bool
	setTestName    bool
}

// assemble stamps grouping numbers in QUERY order, then sorts — the order the
// controllers do it in, and the two are not interchangeable.
//
//	workplanTests = getWorkplanByX(...)      // grouping assigned here
//	filteredTests = filterResultsByLabUnitRoles(...)
//	sortByAccessionAndSequence(filteredTests) // reordered here
//
// So a row's sampleGroupingNumber reflects where its accession first appeared
// in the DATABASE's ordering, while its position in the array reflects Java's
// in-memory sort. On this dataset the two disagree: the database's collation
// ignores punctuation and puts E2E001 first (grouping 1), Java's String
// .compareTo is byte order and puts it last — so the array ends with grouping 1.
func assemble(rows []daoimpl.WorkplanAnalysisRow, opt options) *form.WorkplanForm {
	f := form.NewWorkplanForm()

	items := make([]form.TestResultItemDTO, 0, len(rows))
	grouping := 0
	current := ""
	first := true
	for _, r := range rows {
		item := buildItem(r, opt)

		// The grouping counter advances on a CHANGE of accession, compared
		// against the previous row only — not against every accession seen. A
		// list that revisits an accession later gets a new group for it.
		if first || r.AccessionNumber != current {
			grouping++
			current = r.AccessionNumber
			first = false
		}
		item.SampleGroupingNumber = grouping
		items = append(items, item)
	}

	sortByAccessionAndSequence(items)
	f.WorkplanTests = items
	f.Paging.SearchTermToPage = searchTermToPage(items)
	return &f
}

// sortByAccessionAndSequence ports ResultsLoadUtility's comparator.
//
// The accession comparison is Java's String.compareTo — byte order — NOT the
// database collation the query used. sort.SliceStable, because
// Collections.sort is stable and rows tied on both keys keep their order.
func sortByAccessionAndSequence(items []form.TestResultItemDTO) {
	sort.SliceStable(items, func(i, j int) bool {
		a, b := items[i], items[j]
		if a.AccessionNumber != b.AccessionNumber {
			return a.AccessionNumber < b.AccessionNumber
		}
		// The tie-break reads testSortOrder first and falls back to testName.
		// testSortOrder is never assigned by these four controllers, so the
		// fallback is what applies — and only where testName was set:
		//
		//   ByTest      no testName -> comparator returns 0, query order kept
		//   ByPriority  testName    -> rows sharing an accession sort by NAME,
		//                              which is why GOT/ASAT precedes GPT/ALAT
		//
		// A port that skipped this looks right on ByTest and reorders ByPriority.
		if a.TestName != nil && *a.TestName != "" && b.TestName != nil && *b.TestName != "" {
			return *a.TestName < *b.TestName
		}
		return false
	})
}

// searchTermToPage lists each DISTINCT accession in the sorted result against
// the page it lands on. Every row fits on page 1 here, so the value is always
// "1"; the KEY set and its order are what a port can get wrong.
func searchTermToPage(items []form.TestResultItemDTO) []util.IdValuePair {
	out := []util.IdValuePair{}
	seen := map[string]bool{}
	for _, it := range items {
		if seen[it.AccessionNumber] {
			continue
		}
		seen[it.AccessionNumber] = true
		out = append(out, util.IdValuePair{Id: it.AccessionNumber, Value: strconv.Itoa(1)})
	}
	return out
}

// ErrTypelessSampleItem signals the NPE Java throws from getTestDisplayName.
// Surfaced as a distinct error so the controller can answer with Tomcat's 500
// body rather than a generic failure.
var ErrTypelessSampleItem = errors.New("analysis on a sample item with no type of sample")

// assembleSortedFirst is ByPanel's order of operations: build, sort on the
// accession alone, THEN number. See ByPanel for why the two orders are not
// interchangeable.
func assembleSortedFirst(rows []daoimpl.WorkplanAnalysisRow, opt options) *form.WorkplanForm {
	f := form.NewWorkplanForm()

	items := make([]form.TestResultItemDTO, 0, len(rows))
	for _, r := range rows {
		items = append(items, buildItem(r, opt))
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].AccessionNumber < items[j].AccessionNumber
	})
	grouping := 0
	current := ""
	first := true
	for i := range items {
		if first || items[i].AccessionNumber != current {
			grouping++
			current = items[i].AccessionNumber
			first = false
		}
		items[i].SampleGroupingNumber = grouping
	}
	f.WorkplanTests = items
	f.Paging.SearchTermToPage = searchTermToPage(items)
	return &f
}

// buildItem maps one analysis row onto a workplan row, setting only what the
// controllers set and leaving the 43 constant fields at their Java defaults.
func buildItem(r daoimpl.WorkplanAnalysisRow, opt options) form.TestResultItemDTO {
	item := form.NewTestResultItem()
	item.AccessionNumber = r.AccessionNumber
	if r.ReceivedDate != nil {
		item.ReceivedDate = *r.ReceivedDate
	}
	// setTestId(testType) on ByTest writes the REQUESTED id onto every row;
	// the other three write the analysis's own.
	if opt.forcedTestID != "" {
		item.TestID = opt.forcedTestID
	} else {
		item.TestID = r.TestID
	}
	if opt.setTestName {
		name := r.TestName
		item.TestName = &name
	}
	if opt.setPatientName {
		name := emptyString
		item.PatientName = &name
	}
	return item
}
