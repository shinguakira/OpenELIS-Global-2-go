// Package daoimpl ports org.openelisglobal.login.daoimpl.LoginUserDAOImpl and
// the systemuser lookups the auth path needs. Folder layout mirrors the Java
// source during migration.
package daoimpl

import (
	"errors"

	"gorm.io/gorm"

	"openelis-go/internal/auth/valueholder"
)

// LoginDAOImpl ports LoginUserDAOImpl for the read paths the login flow uses.
type LoginDAOImpl struct {
	DB *gorm.DB
}

// GetByLoginName mirrors LoginUserServiceImpl.getMatch("loginName", name) —
// used by CustomUserDetailsService.loadUserByUsername.
//
// Java's getMatch resolves through Hibernate on the mapped property, i.e. an
// exact, case-SENSITIVE match on login_name; it is `duplicateLoginNameExists`
// (a different, write-path method) that lowercases and trims. Do not
// "harmonize" the two — a case-insensitive login here would be a behavior
// change, not a migration.
//
// Java also treats "more than one row" as not-found: getMatch throws when the
// result is not unique and loadUserByUsername converts that to
// UsernameNotFoundException ("could be duplicates in database or it doesn't
// exist"). Returning (nil, nil) for both reproduces that collapse.
func (d *LoginDAOImpl) GetByLoginName(loginName string) (*valueholder.LoginUser, error) {
	var users []valueholder.LoginUser
	if err := d.DB.Where("login_name = ?", loginName).Limit(2).Find(&users).Error; err != nil {
		return nil, err
	}
	if len(users) != 1 {
		return nil, nil
	}
	return &users[0], nil
}

// PasswordExpiredDayNo mirrors LoginUserDAOImpl.getPasswordExpiredDayNo():
//
//	SELECT floor(current_date - password_expired_dt) * -1 FROM login_user
//	 WHERE login_name = :loginName
//
// i.e. days REMAINING before the password expires (negative once past).
// CustomUserDetailsService treats `<= 0` as credentials-expired.
//
// This is computed in SQL rather than in Go on purpose: `current_date` is the
// DATABASE's date in the database's timezone. Computing it from the Go
// process's clock would diverge whenever the app server and Postgres disagree
// about the day — the same class of bug the b2 site-code work already hit
// (UTC vs host-local). Java gets the DB's answer; so does this.
func (d *LoginDAOImpl) PasswordExpiredDayNo(loginName string) (int, error) {
	var dayNo int
	err := d.DB.Raw(
		"SELECT floor(current_date - password_expired_dt) * -1 FROM clinlims.login_user WHERE login_name = ?",
		loginName,
	).Scan(&dayNo).Error
	if err != nil {
		return 0, err
	}
	return dayNo, nil
}

// SystemUserByLoginName mirrors LoginUserDAOImpl.getSystemUserId():
//
//	SELECT id FROM system_user su WHERE su.login_name = :loginName AND su.is_active='Y'
//
// widened to fetch the name fields too, because the /session DTO needs them and
// Java loads the same row a moment later via SystemUserService.get(id).
//
// Returns (nil, nil) when no ACTIVE row matches — Java's equivalent of
// systemUserId = 0, which makes session setup fail. That is a real, reachable
// state (a login_user whose system_user was deactivated), not an impossible
// one, so it must not be collapsed into an error.
func (d *LoginDAOImpl) SystemUserByLoginName(loginName string) (*valueholder.SystemUser, error) {
	var su valueholder.SystemUser
	err := d.DB.Where("login_name = ? AND is_active = ?", loginName, valueholder.Yes).
		First(&su).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &su, nil
}
