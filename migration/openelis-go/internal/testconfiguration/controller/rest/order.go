package rest

import (
	"encoding/json"
	"net/http"

	authmw "openelis-go/internal/auth/middleware"
	"openelis-go/internal/common/web"
	"openelis-go/internal/testconfiguration/form"
	"openelis-go/internal/testconfiguration/service"
)

// OrderRestController mirrors PanelOrderRestController,
// SampleTypeOrderRestController and TestSectionOrderRestController.
type OrderRestController struct {
	Service *service.OrderService
}

// OrderRoutes registers the three GET/POST pairs.
func OrderRoutes(mux *http.ServeMux, ctrl *OrderRestController) {
	pairs := []struct {
		path string
		get  http.HandlerFunc
	}{
		{"PanelOrder", ctrl.panelForm},
		{"SampleTypeOrder", ctrl.sampleTypeForm},
		{"TestSectionOrder", ctrl.testSectionForm},
	}
	for _, p := range pairs {
		screen := p.path
		web.Register(mux, "GET", "rest/"+p.path, authmw.RequireAdmin(p.get))
		web.Register(mux, "POST", "rest/"+p.path, authmw.RequireAdmin(
			func(w http.ResponseWriter, r *http.Request) {
				var post form.OrderPost
				if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
					web.WriteJSON(w, http.StatusBadRequest, web.ServletError(http.StatusBadRequest))
					return
				}
				f, err := ctrl.Service.Reorder(screen, post, actingUser(r))
				writeForm(w, f, err)
			}))
	}
}

func (c *OrderRestController) panelForm(w http.ResponseWriter, r *http.Request) {
	f, err := c.Service.PanelOrderForm()
	writeForm(w, f, err)
}

func (c *OrderRestController) sampleTypeForm(w http.ResponseWriter, r *http.Request) {
	f, err := c.Service.SampleTypeOrderForm()
	writeForm(w, f, err)
}

func (c *OrderRestController) testSectionForm(w http.ResponseWriter, r *http.Request) {
	f, err := c.Service.TestSectionOrderForm()
	writeForm(w, f, err)
}
