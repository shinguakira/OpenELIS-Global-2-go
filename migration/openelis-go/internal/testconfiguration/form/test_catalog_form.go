// Package form ports org.openelisglobal.testconfiguration.form.
// Folder layout mirrors the Java source during migration.
package form

import testvh "openelis-go/internal/test/valueholder"

// TestCatalogForm mirrors testconfiguration/form/TestCatalogForm.java.
// formName is set to "testCatalogForm" in the Java constructor (BaseForm.setFormName).
type TestCatalogForm struct {
	FormName        string
	TestCatalogList []testvh.TestCatalog
	TestSectionList []string
}
