// Package rest ports the four WorkPlanBy* REST controllers (constitution.md
// Layer IV). Folder layout mirrors the Java source.
package rest

import (
	"net/http"

	"openelis-go/internal/common/web"
	sampleform "openelis-go/internal/sample/form"
	"openelis-go/internal/workplan/daoimpl"
	"openelis-go/internal/workplan/service"
)

// WorkplanRestController serves rest/WorkPlanByTest, ByPanel, ByTestSection
// and ByPriority.
type WorkplanRestController struct {
	Service *service.WorkplanService
}

// Routes registers all four. They are separate paths over one form, so they
// share a controller the way the Java ones share a base class.
func Routes(mux *http.ServeMux, c *WorkplanRestController) {
	web.Register(mux, "GET", "rest/WorkPlanByTest", c.byTest)
	web.Register(mux, "GET", "rest/WorkPlanByPanel", c.byPanel)
	web.Register(mux, "GET", "rest/WorkPlanByTestSection", c.byTestSection)
	web.Register(mux, "GET", "rest/WorkPlanByPriority", c.byPriority)
}

// Each handler mirrors its controller's defaultValue: an absent parameter binds
// to "0" (or "" for priority) and yields the empty form rather than a 400. A
// wrong PARAM NAME therefore reads as "no data" instead of failing loudly —
// the same silent-ignore trap as c2's labNumber, and worth reproducing exactly
// because the screens depend on the empty form rendering.

func (c *WorkplanRestController) byTest(w http.ResponseWriter, r *http.Request) {
	f, err := c.Service.ByTest(param(r, "test_id", "0"))
	c.write(w, f, err)
}

func (c *WorkplanRestController) byPanel(w http.ResponseWriter, r *http.Request) {
	f, err := c.Service.ByPanel(param(r, "panel_id", "0"))
	c.write(w, f, err)
}

func (c *WorkplanRestController) byTestSection(w http.ResponseWriter, r *http.Request) {
	f, err := c.Service.ByTestSection(param(r, "test_section_id", "0"))
	c.write(w, f, err)
}

// byPriority is the only one of the four that can answer 400: `priority` binds
// to the OrderPriority ENUM, so Spring's converter rejects anything outside it
// before the controller runs. An in-enum value with no matching orders is a
// normal empty form.
func (c *WorkplanRestController) byPriority(w http.ResponseWriter, r *http.Request) {
	p := param(r, "priority", "")
	if p != "" && !daoimpl.IsOrderPriority(p) {
		// Same ProblemDetail shape c2 measured on order/dashboard: type and
		// title name MethodArgumentTypeMismatchException while detail names
		// org.springframework.beans.TypeMismatchException.
		web.WriteJSON(w, http.StatusBadRequest, sampleform.TypeMismatchProblem(r))
		return
	}
	f, err := c.Service.ByPriority(p)
	c.write(w, f, err)
}

func (c *WorkplanRestController) write(w http.ResponseWriter, f any, err error) {
	if err != nil {
		// Every failure here answers with Tomcat default error body — the shape
		// an UNHANDLED exception produces, which is what the reproduced NPE is.
		// Spring RFC 7807 ProblemDetail is a separate path, used only for
		// binding failures (see byPriority).
		web.WriteJSON(w, http.StatusInternalServerError, web.ServletError(http.StatusInternalServerError))
		return
	}
	web.WriteJSON(w, http.StatusOK, f)
}

func param(r *http.Request, name, def string) string {
	if !r.URL.Query().Has(name) {
		return def
	}
	return r.URL.Query().Get(name)
}
