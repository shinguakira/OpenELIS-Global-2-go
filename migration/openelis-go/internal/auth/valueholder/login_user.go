// Package valueholder ports org.openelisglobal.login.valueholder (LoginUser,
// UserSessionData) and the slice of org.openelisglobal.systemuser.valueholder
// (SystemUser) the auth path needs. Folder layout mirrors the Java source
// during migration.
//
// Scope note: this is the READ side only. Nothing in internal/auth writes to
// login_user / system_user / system_user_role — no user creation, no password
// change, no lockout administration (see migration/auth-adoption-plan.md §1).
package valueholder

import "time"

// LoginUser mirrors login.valueholder.LoginUser. Maps to clinlims.login_user.
//
// Column types come from the live schema, not from the JPA annotations:
// account_locked / account_disabled / is_admin are varchar(1) holding 'Y'/'N'
// (not booleans), and user_time_out is a varchar(3) holding MINUTES as text.
type LoginUser struct {
	ID        int64  `gorm:"column:id"`
	LoginName string `gorm:"column:login_name"`
	// Password is a bcrypt digest: '$2a$', cost 12, 60 chars. Java uses a plain
	// BCryptPasswordEncoder (SecurityConfig.passwordEncoder) — NOT a
	// DelegatingPasswordEncoder — so there is no "{id}" prefix and no
	// plaintext/legacy fallback to reproduce.
	Password string `gorm:"column:password"`
	// PasswordExpiredDT drives the credentials-expired check. Java never
	// compares the date directly: it computes days-remaining in SQL (see
	// LoginDAO.PasswordExpiredDayNo) and treats <= 0 as expired.
	PasswordExpiredDT time.Time `gorm:"column:password_expired_dt"`
	AccountLocked     string    `gorm:"column:account_locked"`
	AccountDisabled   string    `gorm:"column:account_disabled"`
	IsAdmin           string    `gorm:"column:is_admin"`
	UserTimeOut       string    `gorm:"column:user_time_out"`
}

// TableName pins the schema-qualified table (GORM would otherwise guess
// "login_users").
func (LoginUser) TableName() string { return "clinlims.login_user" }

// Yes mirrors IActionConstants.YES. Java compares with equalsIgnoreCase, so
// this port does too rather than testing == "Y".
const Yes = "Y"

// SystemUser mirrors systemuser.valueholder.SystemUser for the fields the auth
// path reads. Maps to clinlims.system_user.
//
// The join to LoginUser is by login_name STRING and requires is_active='Y'
// (LoginUserDAOImpl.getSystemUserId) — there is no foreign key between the two
// tables. Do not "improve" this into an id join: a login_user whose name has no
// ACTIVE system_user row is a real, reachable state that Java handles
// distinctly.
type SystemUser struct {
	ID        int64  `gorm:"column:id"`
	LoginName string `gorm:"column:login_name"`
	FirstName string `gorm:"column:first_name"`
	LastName  string `gorm:"column:last_name"`
	IsActive  string `gorm:"column:is_active"`
}

// TableName pins the schema-qualified table.
func (SystemUser) TableName() string { return "clinlims.system_user" }
