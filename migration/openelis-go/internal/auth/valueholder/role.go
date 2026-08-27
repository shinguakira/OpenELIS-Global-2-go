package valueholder

// Role mirrors role.valueholder.Role for the fields the auth path reads. Maps
// to clinlims.system_role.
//
// TRAP (verified live, migration/auth-adoption-plan.md §6.1): system_role.name
// is `character(30)` — PostgreSQL blank-pads it, so a driver scanning the raw
// column gets "Reception" followed by 21 spaces. Casual `psql` inspection hides
// this because `'[' || name || ']'` casts bpchar to text, which strips the
// padding; `SELECT name` does not. Java trims at every comparison site; this
// port trims ONCE, in the DAO (RoleDAOImpl), so nothing downstream ever sees a
// padded name.
//
// Consequence if the trim is dropped: every `role == "Reception"` check
// silently returns false and every role-gated endpoint 403s for everyone. The
// e2e spec p0-auth.spec.ts asserts the DB column IS padded before asserting the
// API value is trimmed, so removing the trim fails a test rather than
// degrading quietly.
type Role struct {
	ID   int64  `gorm:"column:id"`
	Name string `gorm:"column:name"`
}

// TableName pins the schema-qualified table.
func (Role) TableName() string { return "clinlims.system_role" }
