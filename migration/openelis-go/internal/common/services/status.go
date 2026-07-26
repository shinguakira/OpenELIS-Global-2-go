// Package services ports org.openelisglobal.common.services (the bits the ported
// endpoints need). Folder layout mirrors the Java source during migration.
package services

import "database/sql"

// StatusService mirrors StatusService for status-type lookups: resolve a
// status_of_sample id from its (status_type, internal name). Java builds these
// maps once at @PostConstruct; we load them once at construction too.
type StatusService struct {
	idByKey map[string]string // status_type + "\x00" + name -> id
}

// NewStatusService loads the status_of_sample rows into an in-memory map.
func NewStatusService(db *sql.DB) (*StatusService, error) {
	rows, err := db.Query(`SELECT id::text, status_type, name FROM clinlims.status_of_sample`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := map[string]string{}
	for rows.Next() {
		var id, statusType, name string
		if err := rows.Scan(&id, &statusType, &name); err != nil {
			return nil, err
		}
		m[statusType+"\x00"+name] = id
	}
	return &StatusService{idByKey: m}, rows.Err()
}

// IDByName returns the status id for (statusType, internal name), or "" if absent.
func (s *StatusService) IDByName(statusType, name string) string {
	return s.idByKey[statusType+"\x00"+name]
}
