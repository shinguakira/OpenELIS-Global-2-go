// Package services ports org.openelisglobal.common.services (the bits the ported
// endpoints need). Folder layout mirrors the Java source during migration.
package services

import (
	"strconv"

	"openelis-go/internal/common/daoimpl"
)

type statusEntry struct {
	id    string
	label string
}

// StatusService mirrors StatusService for status-type lookups: resolve a
// status_of_sample id and localized label from its (status_type, internal name).
// Java builds these maps once at @PostConstruct; we load them once at construction.
// Holds no DB handle — all access goes through StatusDAOImpl.
type StatusService struct {
	entryByKey map[string]statusEntry // status_type + "\x00" + name → {id, label}
}

// NewStatusService loads the status_of_sample rows (via the DAO) into an
// in-memory map. msgs is the parsed message_en.properties (from i18n.Messages());
// it is used to resolve each row's display_key into a human-readable label —
// mirroring Java's BaseObject.getLocalizedName() → MessageUtil.getContextualMessage(nameKey).
// Falls back to the DB name column when display_key is empty or the key is absent
// (same as Java's getDefaultLocalizedName() fallback path).
func NewStatusService(dao *daoimpl.StatusDAOImpl, msgs map[string]string) (*StatusService, error) {
	rows, err := dao.GetAll()
	if err != nil {
		return nil, err
	}

	m := map[string]statusEntry{}
	for _, r := range rows {
		label := msgs[r.DisplayKey]
		if label == "" {
			label = r.Name
		}
		m[r.StatusType+"\x00"+r.Name] = statusEntry{id: strconv.FormatInt(r.ID, 10), label: label}
	}
	return &StatusService{entryByKey: m}, nil
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
