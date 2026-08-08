// Package rest ports org.openelisglobal.localization.controller.rest — the
// SupportedLocale REST controller (@RequestMapping("/rest/supportedlocales")).
// Folder layout mirrors the Java source during migration.
package rest

import (
	"net/http"
	"strconv"

	"openelis-go/internal/common/web"
	"openelis-go/internal/localization/service"
	"openelis-go/internal/localization/valueholder"
)

// supportedLocaleDTO mirrors SupportedLocaleRestController.SupportedLocaleDTO.
// Field order matches the Java DTO's JSON output.
type supportedLocaleDTO struct {
	Id          string `json:"id"`
	LocaleCode  string `json:"localeCode"`
	DisplayName string `json:"displayName"`
	Active      bool   `json:"active"`
	Fallback    bool   `json:"fallback"`
	SortOrder   int    `json:"sortOrder"`
}

func toDTO(l valueholder.SupportedLocale) supportedLocaleDTO {
	return supportedLocaleDTO{
		Id:          strconv.FormatInt(l.Id, 10),
		LocaleCode:  l.LocaleCode,
		DisplayName: l.DisplayName,
		Active:      l.Active,
		Fallback:    l.Fallback,
		SortOrder:   l.SortOrder,
	}
}

func toDTOs(list []valueholder.SupportedLocale) []supportedLocaleDTO {
	dtos := make([]supportedLocaleDTO, 0, len(list)) // non-nil → serializes as []
	for _, l := range list {
		dtos = append(dtos, toDTO(l))
	}
	return dtos
}

// Routes registers the supportedlocales endpoints. Mirrors the @GetMapping methods
// on SupportedLocaleRestController. NOTE: when /{id} is later ported, the router
// must match literal /fallback before the /{id} pattern (as Spring does).
func Routes(mux *http.ServeMux, svc *service.SupportedLocaleService) {
	web.Register(mux, "GET", "rest/supportedlocales", allLocales(svc))
	web.Register(mux, "GET", "rest/supportedlocales/active", activeLocales(svc))
	web.Register(mux, "GET", "rest/supportedlocales/fallback", fallbackLocale(svc))
}

// allLocales reproduces getAllLocales() — GET /rest/supportedlocales.
func allLocales(svc *service.SupportedLocaleService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := svc.GetAll()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, toDTOs(list))
	}
}

// activeLocales reproduces getActiveLocales() — GET /rest/supportedlocales/active
// (WHERE is_active ORDER BY sort_order ASC).
func activeLocales(svc *service.SupportedLocaleService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := svc.GetAllActive()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, toDTOs(list))
	}
}

// fallbackLocale reproduces getFallbackLocale() — GET /rest/supportedlocales/fallback.
// Returns a single object, or 404 when no fallback row exists (Java: notFound()).
func fallbackLocale(svc *service.SupportedLocaleService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fb, err := svc.GetFallback()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if fb == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		web.WriteJSON(w, http.StatusOK, toDTO(*fb))
	}
}
