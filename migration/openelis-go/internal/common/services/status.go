// Package services ports org.openelisglobal.common.services (the bits the ported
// endpoints need). Folder layout mirrors the Java source during migration.
package services

import "database/sql"

type statusEntry struct {
	id    string
	label string
}

// StatusService mirrors StatusService for status-type lookups: resolve a
// status_of_sample id and localized label from its (status_type, internal name).
// Java builds these maps once at @PostConstruct; we load them once at construction.
type StatusService struct {
	entryByKey map[string]statusEntry // status_type + "\x00" + name → {id, label}
}

// NewStatusService loads the status_of_sample rows into an in-memory map.
// msgs is the parsed message_en.properties (from i18n.Messages()); it is used
// to resolve each row's display_key into a human-readable label — mirroring
// Java's BaseObject.getLocalizedName() → MessageUtil.getContextualMessage(nameKey).
// Falls back to the DB name column when display_key is empty or the key is absent
// (same as Java's getDefaultLocalizedName() fallback path).
func NewStatusService(db *sql.DB, msgs map[string]string) (*StatusService, error) {
	rows, err := db.Query(`
		SELECT id::text, status_type, name, COALESCE(display_key, '')
		FROM clinlims.status_of_sample`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := map[string]statusEntry{}
	for rows.Next() {
		var id, statusType, name, displayKey string
		if err := rows.Scan(&id, &statusType, &name, &displayKey); err != nil {
			return nil, err
		}
		label := msgs[displayKey]
		if label == "" {
			label = name
		}
		m[statusType+"\x00"+name] = statusEntry{id: id, label: label}
	}
	return &StatusService{entryByKey: m}, rows.Err()
}

// IDByName returns the status id for (statusType, internal name).
// Returns "-1" if absent — matching Java's StatusService.getStatusID behaviour
// (returns "-1" when the enum→row map has no entry for the given status).
func (s *StatusService) IDByName(statusType, name string) string {
	if e, ok := s.entryByKey[statusType+"\x00"+name]; ok {
		return e.id
	}
	return "-1"
}

// LabelByName returns the localized label for (statusType, internal name),
// resolved via display_key → message_en.properties. Returns "" if absent.
func (s *StatusService) LabelByName(statusType, name string) string {
	return s.entryByKey[statusType+"\x00"+name].label
}
