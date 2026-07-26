// Package rest ports org.openelisglobal.testcalculated.controller.rest — the
// test-calculation domain's REST controllers. Folder layout mirrors the Java
// source (testcalculated/controller/rest) during migration.
package rest

import (
	"net/http"

	"openelis-go/internal/common/web"
	"openelis-go/internal/testcalculated/valueholder"
)

// Routes registers the test-calculation REST endpoints. Mirrors the @GetMapping
// methods on CalculatedValueRestController (@RequestMapping("/rest/")).
func Routes(mux *http.ServeMux) {
	web.Register(mux, "GET", "rest/math-functions", MathFunctions)
}

// MathFunctions reproduces CalculatedValueRestController#getMathFunctions
// (CalculatedValueRestController.java:111 -> Operation.mathFunctions()):
//
//	GET /rest/math-functions -> [{"id","value"}, ...]  (the 14-operator catalog)
func MathFunctions(w http.ResponseWriter, r *http.Request) {
	web.WriteJSON(w, http.StatusOK, valueholder.MathFunctions())
}
