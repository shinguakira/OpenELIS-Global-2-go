// Package valueholder ports org.openelisglobal.test.valueholder.
// Folder layout mirrors the Java source during migration.
package valueholder

import locvh "openelis-go/internal/localization/valueholder"

// TestCatalog mirrors test/valueholder/TestCatalog.java — the read-only projection
// used by TestCatalogRestController.showTestCatalog() to build the catalog list.
// Fields present here are those the endpoint serializes; write-side fields
// (resultLimits, dictionaryValues, etc.) are omitted until write endpoints are ported.
type TestCatalog struct {
	ID                  string
	Localization        locvh.Localization
	TestUnit            string
	SampleType          string
	Panel               string
	ResultType          string
	Active              string
	Orderable           string
	Loinc               *string
	Uom                 string
	SignificantDigits   string
	HasLimitValues      bool
	HasDictionaryValues bool
	// Sort key used internally by TestCatalogRestController.createTestList();
	// not serialized to the response.
	TestSortOrder int64
}
