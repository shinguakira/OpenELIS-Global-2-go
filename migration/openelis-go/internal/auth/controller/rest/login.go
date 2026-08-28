// Package rest ports the auth endpoints. In Java these live in three different
// places — /ValidateLogin is intercepted by Spring Security's login FILTER
// (never reaching a controller), /session is a @GetMapping on
// LoginPageController, and /Logout is the LogoutFilter — so there is no single
// Java class this file mirrors. It reproduces their combined observable
// contract.
//
// Per constitution.md Layer IV: request/response mapping and service calls
// only. The DTO lives in internal/auth/form (Layer V), the decision in
// internal/auth/service (Layer III).
package rest

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"openelis-go/internal/auth/csrf"
	"openelis-go/internal/auth/form"
	"openelis-go/internal/auth/middleware"
	"openelis-go/internal/auth/service"
	"openelis-go/internal/auth/session"
	"openelis-go/internal/common/web"
)

// LoginRestController serves the three auth endpoints.
type LoginRestController struct {
	Service *service.AuthService
	Store   session.Store
}

// Routes registers the auth endpoints. All three are registered OPEN
// (unauthenticated) because they are Java's LOGIN_PAGES — SecurityConfig's
// `.requestMatchers(LOGIN_PAGES).permitAll()` covers exactly
// /LoginPage, /ValidateLogin and /session. Logout is reachable while
// authenticated by definition, and carries its own CSRF check.
func Routes(mux *http.ServeMux, ctrl *LoginRestController) {
	// POST ValidateLogin — Spring's UsernamePasswordAuthenticationFilter with
	// usernameParameter("loginName") / passwordParameter("password").
	// CSRF-exempt: SecurityConfig does `.csrf(c -> c.ignoringRequestMatchers
	// ("/ValidateLogin"))`.
	web.RegisterOpen(mux, "POST", "ValidateLogin", ctrl.validateLogin)

	// GET session — LoginPageController.getSesssionDetails. In LOGIN_PAGES, so
	// anonymous callers get the bootstrap shape rather than a redirect.
	web.RegisterOpen(mux, "GET", "session", ctrl.sessionDetails)

	// POST Logout — Spring's LogoutFilter:
	// .logout(l -> l.logoutUrl("/Logout").logoutSuccessUrl("/LoginPage")
	//              .invalidateHttpSession(true))
	web.RegisterOpen(mux, "POST", "Logout", ctrl.logout)
}

// validateLogin reproduces the login filter plus both custom handlers.
func (c *LoginRestController) validateLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeLoginError(w, service.ErrGeneric)
		return
	}
	loginName := r.PostFormValue("loginName")
	password := r.PostFormValue("password")

	principal, ttl, err := c.Service.Authenticate(loginName, password)
	if err != nil {
		var authErr *service.AuthError
		switch {
		case errors.As(err, &authErr):
			writeLoginError(w, authErr)
		case errors.Is(err, service.ErrNoOeUser):
			// Credentials were fine; session setup failed. Java redirects here
			// even for ?apiCall=true — the JSON contract does not cover this
			// branch. Verified live.
			session.Redirect(w, r, "/LoginPage")
		default:
			log.Printf("auth: login failed for %q: %v", loginName, err)
			writeLoginError(w, service.ErrGeneric)
		}
		return
	}

	// sessionFixation().migrateSession(): drop whatever id the client presented
	// and issue a fresh one.
	if old := session.IDFromRequest(r); old != "" {
		c.Store.Delete(old)
	}
	id, err := c.Store.New(principal, ttl)
	if err != nil {
		log.Printf("auth: session creation failed for %q: %v", loginName, err)
		writeLoginError(w, service.ErrGeneric)
		return
	}
	session.SetCookie(w, r, id)

	// CustomFormAuthenticationSuccessHandler.handleApiLogin writes exactly
	// {"success":true} with content-type application/json — no redirect, no
	// other field.
	//
	// Divergence, deliberate: Java branches on ?apiCall=true and otherwise does
	// a SavedRequestAware redirect for browser form posts. The Go service has
	// no server-rendered pages to redirect to (the React frontend and the e2e
	// suite both use apiCall=true), so it always answers JSON. Pinning the
	// browser branch would mean porting the JSP login flow, which is out of
	// scope — auth-adoption-plan.md §1.
	web.WriteJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// writeLoginError reproduces CustomAuthenticationFailureHandler.handleApiLogin:
// HTTP 401, application/json, body {"error":"<key>"} and nothing else.
func writeLoginError(w http.ResponseWriter, e *service.AuthError) {
	web.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": e.Key})
}

// sessionDetails reproduces LoginPageController.getSesssionDetails.
func (c *LoginRestController) sessionDetails(w http.ResponseWriter, r *http.Request) {
	id := session.IDFromRequest(r)

	if id == "" {
		// No cookie at all: create a session so the client has an id, exactly
		// as request.getSession() does in Java (which creates on demand). It
		// carries a nil principal — an id is not an identity.
		newID, err := c.Store.New(nil, service.DefaultSessionTimeout)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		session.SetCookie(w, r, newID)
		web.WriteJSON(w, http.StatusOK, form.UserSession{
			Authenticated: false,
			SessionID:     newID,
		})
		return
	}

	principal, ok := c.Store.Get(id)
	if !ok {
		// A dead or unknown id — e.g. the client still holds the JSESSIONID it
		// had before logging out. Java's
		// sessionManagement().invalidSessionUrl("/LoginPage") intercepts BEFORE
		// the permitAll rule, so even /session redirects rather than reporting
		// authenticated:false. Verified live.
		//
		// The cookie is deliberately NOT cleared: Spring's LogoutFilter is not
		// configured with deleteCookies("JSESSIONID"), so a real client keeps
		// presenting the dead id and keeps getting this redirect. Clearing it
		// here would turn the next call into a 200, diverging from Java.
		session.Redirect(w, r, "/LoginPage")
		return
	}
	if principal == nil {
		// An existing but anonymous session: same id, still not authenticated.
		web.WriteJSON(w, http.StatusOK, form.UserSession{
			Authenticated: false,
			SessionID:     id,
		})
		return
	}

	// Fresh mask on EVERY read — see auth/csrf. Handing out the same string
	// twice would fail the parity oracle and defeat the masking's purpose.
	masked, err := csrf.Mask(principal.CSRFToken)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	roles := append([]string{}, principal.Roles...)
	web.WriteJSON(w, http.StatusOK, form.UserSession{
		Authenticated: true,
		LoginMethod:   "FORM",
		SessionID:     id,
		UserID:        strconv.FormatInt(principal.SystemUserID, 10),
		LoginName:     principal.LoginName,
		FirstName:     principal.FirstName,
		LastName:      principal.LastName,
		Roles:         &roles,
		CSRF:          masked,
	})
}

// logout reproduces the LogoutFilter — including the fact that it sits BEHIND
// the CsrfFilter, so a logout without a valid token is refused and the session
// survives. Verified live on Java: no/forged token → 302 to
// <context>/Home?access=denied with the session intact; valid token → 302 to
// <context>/LoginPage with the session gone.
func (c *LoginRestController) logout(w http.ResponseWriter, r *http.Request) {
	id := session.IDFromRequest(r)
	principal, ok := c.Store.Get(id)
	if !ok || principal == nil {
		// Nothing to log out of. Java's invalidSessionUrl handles the
		// stale-cookie case the same way.
		session.Redirect(w, r, "/LoginPage")
		return
	}
	if !csrf.Valid(principal.CSRFToken, submittedCSRF(r)) {
		middleware.DenyAccess(w, r, "CSRF token missing or invalid")
		return
	}
	c.Store.Delete(id)
	// The cookie is NOT cleared — see sessionDetails. Java leaves the client
	// holding a dead id on purpose, and the next request's redirect depends on
	// it still being sent.
	session.Redirect(w, r, "/LoginPage")
}

func submittedCSRF(r *http.Request) string {
	if v := r.Header.Get("X-CSRF-TOKEN"); v != "" {
		return v
	}
	if err := r.ParseForm(); err == nil {
		return r.PostFormValue("_csrf")
	}
	return ""
}
