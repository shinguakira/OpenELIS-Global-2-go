// Package form holds the DTOs the auth controller returns — constitution.md
// Layer V. It mirrors org.openelisglobal.login.bean.UserSession.
package form

// UserSession mirrors login.bean.UserSession as serialized by Java's globally
// configured ObjectMapper (AppConfig:182 — Include.NON_NULL), so every
// omitempty below corresponds to a field Jackson drops when null.
//
// Verified live, anonymous:
//
//	{"authenticated":false,"sessionId":"…"}
//
// Verified live, authenticated (a user with no lab-unit roles):
//
//	{"authenticated":true,"loginMethod":"FORM","sessionId":"…","userId":"9901",
//	 "loginName":"e2e_reception","firstName":"E2E","lastName":"Reception",
//	 "roles":["Reception"],"csrf":"…"}
//
// Field order matters for byte-identical responses: Jackson emits declaration
// order, so this struct preserves UserSession.java's.
type UserSession struct {
	// Authenticated is `userSessionData != null` in Java
	// (UserModuleServiceImpl.isSessionExpired) — NOT "has a SecurityContext".
	// Merging both notions into one principal keeps them from diverging.
	Authenticated bool `json:"authenticated"`
	// LoginMethod is always "FORM" here: SAML/OAUTH/CERT are separate Java
	// filter chains, all disabled by default and out of scope for this port
	// (auth-adoption-plan.md §1, §9.4).
	LoginMethod string `json:"loginMethod,omitempty"`
	// SessionId is emitted even when unauthenticated — verified live.
	SessionID string `json:"sessionId"`
	// UserId is system_user.id as a STRING (Java's SystemUser.getId()).
	UserID    string `json:"userId,omitempty"`
	LoginName string `json:"loginName,omitempty"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	// Roles are TRIMMED system_role.name values.
	//
	// A POINTER to the slice, not a bare slice, and the distinction is load
	// bearing: Jackson's Include.NON_NULL drops only nulls, so an authenticated
	// user with zero grants gets `"roles":[]` (verified live) while an
	// anonymous session omits the key entirely. Go's `omitempty` cannot express
	// that — it drops empty slices too — so nil means "omit" and a pointer to
	// an empty slice means "emit []".
	Roles *[]string `json:"roles,omitempty"`
	// UserLabRolesMap comes from the lab-unit-roles subsystem
	// (user_lab_unit_roles), which is NOT ported in P0: no ported endpoint
	// reads it, and it is absent from the response for every user without such
	// a grant. Kept in the DTO so the field name is already right when that
	// subsystem lands.
	UserLabRolesMap map[string][]string `json:"userLabRolesMap,omitempty"`
	// CSRF is the session token, XOR-MASKED afresh on every read (see
	// auth/csrf). Jackson renders the getCSRF() property as "csrf".
	CSRF string `json:"csrf,omitempty"`
	// LoginLabUnit is set only when REQUIRE_LAB_UNIT_AT_LOGIN is on and the
	// user picked a unit; not ported for the same reason as UserLabRolesMap.
	LoginLabUnit string `json:"loginLabUnit,omitempty"`
}
