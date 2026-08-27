// Package service ports the authentication decision that Spring Security makes
// across CustomUserDetailsService, DaoAuthenticationProvider and
// CustomFormAuthenticationSuccessHandler. Per constitution.md Layer III, all
// data is compiled here so the controller only shapes a response.
package service

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"openelis-go/internal/auth/csrf"
	"openelis-go/internal/auth/daoimpl"
	"openelis-go/internal/auth/session"
	"openelis-go/internal/auth/valueholder"
)

// DefaultSessionTimeoutMinutes mirrors
// CustomFormAuthenticationSuccessHandler.DEFAULT_SESSION_TIMEOUT_IN_MINUTES.
const DefaultSessionTimeoutMinutes = 20

// DefaultSessionTimeout is the same value as a Duration, for the anonymous
// bootstrap session GET /session creates on demand.
const DefaultSessionTimeout = DefaultSessionTimeoutMinutes * time.Minute

// AuthError carries the i18n key Java's CustomAuthenticationFailureHandler
// emits for an apiCall login. The five keys below are the complete set; the
// unknown-user and wrong-password cases deliberately share one so the API
// cannot be used to enumerate users.
type AuthError struct{ Key string }

func (e *AuthError) Error() string { return e.Key }

// The five failure keys, verified live against Java.
var (
	ErrInvalidCredentials = &AuthError{Key: "error.invalidcredentials"}
	ErrExpiredCredentials = &AuthError{Key: "error.expiredCredentials"}
	ErrDisabledCredential = &AuthError{Key: "error.disabledCredentials"}
	ErrLockedCredentials  = &AuthError{Key: "error.lockedCredentials"}
	ErrGeneric            = &AuthError{Key: "error.generic"}
)

// ErrNoOeUser means the credentials were correct but no ACTIVE system_user
// carries the login name, so session setup cannot complete.
//
// This is NOT one of the five JSON keys: in Java the credential check has
// already succeeded by this point, and the failure happens inside
// CustomFormAuthenticationSuccessHandler.setupUserSession, which redirects to
// /LoginPage — escaping the apiCall JSON contract entirely, even for
// ?apiCall=true. Verified live. The controller reproduces the redirect.
var ErrNoOeUser = errors.New("no active system_user for login name")

// AuthService ports the login decision.
type AuthService struct {
	LoginDAO  *daoimpl.LoginDAOImpl
	RoleDAO   *daoimpl.RoleDAOImpl
	ModuleDAO *daoimpl.ModuleDAOImpl
}

// Authenticate reproduces Spring's credential check, INCLUDING ITS ORDER.
//
// AbstractUserDetailsAuthenticationProvider runs:
//  1. preAuthenticationChecks  — locked, then disabled, then account-expired
//     (CustomUserDetailsService hardcodes accountNonExpired=true, so that third
//     check can never fire and is not ported)
//  2. additionalAuthenticationChecks — the bcrypt comparison
//  3. postAuthenticationChecks — credentials-expired
//
// The order is observable and is pinned by two e2e tests: a LOCKED account with
// a WRONG password still reports error.lockedCredentials (state first), while
// an EXPIRED-password account with a WRONG password reports
// error.invalidcredentials (password before the post-check). Reordering these
// fails p0-auth.spec.ts rather than degrading quietly.
func (s *AuthService) Authenticate(loginName, password string) (*session.Principal, time.Duration, error) {
	user, err := s.LoginDAO.GetByLoginName(loginName)
	if err != nil {
		return nil, 0, err
	}
	// Java's UsernameNotFoundException is converted to BadCredentialsException
	// by AbstractUserDetailsAuthenticationProvider (hideUserNotFoundExceptions
	// defaults to true), and both map to the same key regardless. Keep them
	// identical: a distinct key here would be a user-enumeration oracle.
	if user == nil {
		// Still spend the bcrypt time so a missing user is not distinguishable
		// by response latency. Java does not do this (it short-circuits), but
		// the timing side channel is invisible to the parity oracle and closing
		// it cannot change any asserted behavior.
		_ = bcrypt.CompareHashAndPassword(
			[]byte("$2a$12$fXxEjo/QbHU7NVgpjrwfLOJFtBtJoF7tZ3xsv580buiMg5OaDkjpG"),
			[]byte(password))
		return nil, 0, ErrInvalidCredentials
	}

	// 1. preAuthenticationChecks — locked BEFORE disabled, matching
	//    AbstractUserDetailsAuthenticationProvider.DefaultPreAuthenticationChecks.
	if strings.EqualFold(user.AccountLocked, valueholder.Yes) {
		return nil, 0, ErrLockedCredentials
	}
	if strings.EqualFold(user.AccountDisabled, valueholder.Yes) {
		return nil, 0, ErrDisabledCredential
	}

	// 2. the password itself
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, 0, ErrInvalidCredentials
	}

	// 3. postAuthenticationChecks — credentials-expired, AFTER the password.
	dayNo, err := s.LoginDAO.PasswordExpiredDayNo(user.LoginName)
	if err != nil {
		return nil, 0, err
	}
	if dayNo <= 0 {
		return nil, 0, ErrExpiredCredentials
	}

	// Credentials accepted. Everything below mirrors
	// CustomFormAuthenticationSuccessHandler.setupUserSession — which runs
	// AFTER authentication succeeded, so its failures look different (redirect,
	// not JSON).
	su, err := s.LoginDAO.SystemUserByLoginName(user.LoginName)
	if err != nil {
		return nil, 0, err
	}
	if su == nil {
		return nil, 0, ErrNoOeUser
	}

	roles, err := s.RoleDAO.RolesForUser(su.ID)
	if err != nil {
		return nil, 0, err
	}

	// The permitted-module set is computed HERE, once, exactly as
	// setupUserSession does before stashing it in the session as
	// PERMITTED_ACTIONS_MAP. Computing it per request instead would be a
	// behavior change: in Java a role granted mid-session has no effect until
	// the user logs in again.
	//
	// Java only populates it when permissions.agent is Role — which it is (see
	// AuthzService.PermissionsAgentRole); the other mode is not ported.
	modules, err := s.ModuleDAO.PermittedModulesForUser(su.ID)
	if err != nil {
		return nil, 0, err
	}

	token, err := csrf.NewToken()
	if err != nil {
		return nil, 0, err
	}

	p := &session.Principal{
		SystemUserID: su.ID,
		LoginName:    user.LoginName,
		FirstName:    su.FirstName,
		LastName:     su.LastName,
		IsAdmin:      strings.EqualFold(user.IsAdmin, valueholder.Yes),
		Roles:        roles,
		Modules:      modules,
		CSRFToken:    token,
	}
	return p, sessionTTL(user.UserTimeOut), nil
}

// sessionTTL mirrors setupUserSession: login_user.user_time_out is MINUTES held
// as text, applied as HttpSession.setMaxInactiveInterval(minutes * 60) — i.e.
// an idle timeout, which is why session.MemoryStore slides the expiry on read.
//
// Divergence, deliberate and unreachable in practice: a non-numeric
// user_time_out makes Java throw NumberFormatException inside the success
// handler and redirect to /LoginPage. The column is varchar(3) written only by
// the application, so no such row exists; falling back to the documented
// default is safer than reproducing a crash for a value that cannot occur.
//
// NOT ported (and deliberately so): Java also stores
// UserSessionData.userTimeOut as minutes*3600 while the REAL session timeout is
// minutes*60 — a genuine Java bug. Nothing in the ported surface reads that
// field, so there is nothing to pin; see auth-adoption-plan.md §10.
func sessionTTL(userTimeOut string) time.Duration {
	minutes, err := strconv.Atoi(strings.TrimSpace(userTimeOut))
	if err != nil || minutes <= 0 {
		minutes = DefaultSessionTimeoutMinutes
	}
	return time.Duration(minutes) * time.Minute
}
