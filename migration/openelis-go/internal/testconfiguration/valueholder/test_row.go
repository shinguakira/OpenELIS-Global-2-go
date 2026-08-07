// Package valueholder holds the DB projections for the test-catalog read path.
// Folder layout mirrors the Java source during migration.
package valueholder

// TestRow is the DB projection for one test in the catalog read path — the
// flattened result of joining test → test_section → type_of_sample → unit_of_measure.
// Java gets this shape implicitly via Hibernate lazy-loading off the Test entity;
// Go materialises it as one explicit projection to avoid the N+1 the Java
// controller incurs in createTestList().
//
// Fields are exported so GORM's Scan can populate them via reflection.
// Nullable columns use pointers (nil = SQL NULL).
type TestRow struct {
	TestID      string  `gorm:"column:test_id"`
	SectionName string  `gorm:"column:section_name"`
	SortOrder   int64   `gorm:"column:sort_order"`
	LocID       *string `gorm:"column:name_localization_id"`
	SampleType  string  `gorm:"column:sample_type"`
	Active      string  `gorm:"column:is_active"`
	Orderable   string  `gorm:"column:orderable"`
	Loinc       *string `gorm:"column:loinc"`
	UomName     string  `gorm:"column:uom_name"`
}
