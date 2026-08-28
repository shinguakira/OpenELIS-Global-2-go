// Package daoimpl ports org.openelisglobal.siteinformation.daoimpl.
// Folder layout mirrors the Java source during migration.
package daoimpl

import (
	"errors"
	"strconv"
	"time"

	"gorm.io/gorm"

	"openelis-go/internal/common/audittrail"
	"openelis-go/internal/siteinformation/valueholder"
)

// siteInformationTable is the reference_tables name the audit rows key on.
const siteInformationTable = "site_information"

// ErrNoSuchRow is returned when an update names an id that does not exist.
//
// Java has no such error: siteInformationService.get() returns null and the
// next line dereferences it, so the request ends as a 500. The controller turns
// this into that same 500 — the value exists so the DAO does not have to
// pretend the update succeeded.
var ErrNoSuchRow = errors.New("site_information: no such row")

// SiteInformationDAOImpl ports SiteInformationDAOImpl.
type SiteInformationDAOImpl struct {
	DB    *gorm.DB
	Audit *audittrail.Service
}

// writeTime is the timestamp a write stamps on both the row and its audit row.
//
// Truncated to MILLISECONDS because Java builds it with
// new Timestamp(System.currentTimeMillis()) — letting the database supply now()
// would store microseconds, and the delete payload renders the value back out
// as text, where three digits against six is a visible difference.
func writeTime() time.Time { return time.Now().UTC().Truncate(time.Millisecond) }

const selectColumns = `si.id, si.name, si.lastupdated, si.description, si.value,
	si.encrypted, si.domain_id, si.value_type, si.instruction_key, si."group",
	si.schedule_id, si.tag, si.dictionary_category_id, si.description_key, si.name_key,
	d.name AS domain_name, d.description AS domain_description`

func (dao *SiteInformationDAOImpl) base() *gorm.DB {
	return dao.DB.Table("clinlims.site_information AS si").
		Select(selectColumns).
		Joins("LEFT JOIN clinlims.site_information_domain AS d ON d.id = si.domain_id")
}

// Get returns one row by id, or nil when it does not exist.
//
// Java has no such nil: siteInformationService.get(id) hands back null and the
// caller dereferences it, which is why an unknown ID answers 500 rather than
// 404. The controller reproduces that; the DAO stays honest.
func (dao *SiteInformationDAOImpl) Get(id string) (*valueholder.SiteInformation, error) {
	var row valueholder.SiteInformation
	err := dao.base().Where("si.id = ?", id).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// ByDomainName ports getSiteInformationByDomainName:
//
//	From SiteInformation si where si.domain.name = :domainName order by si.name
//
// The ORDER BY is applied by the DATABASE, so it sorts under the database's own
// collation — case-insensitive here, which puts allowLanguageChange before
// bannerHeading before BarCodeType. That is the opposite of the c3 orderings,
// where Java sorted in memory with String.compareTo and the port needed
// COLLATE "C" to reproduce byte order. Measured both ways against the live
// response; the two differ on this data and the DB collation is the match.
func (dao *SiteInformationDAOImpl) ByDomainName(domainName string) ([]valueholder.SiteInformation, error) {
	rows := []valueholder.SiteInformation{}
	err := dao.base().Where("d.name = ?", domainName).Order("si.name").Scan(&rows).Error
	return rows, err
}

// DictionaryEntriesByCategory ports
// dictionaryService.getDictionaryEntriesByCategoryId, which reaches
// BaseDAOImpl.getAllMatchingOrdered with an EMPTY order list — so Java applies
// no ORDER BY at all and the row order is whatever the plan yields.
//
// It is not arbitrary in practice: the planner serves the query from the unique
// index on (dictionary_category_id, dict_entry), so the scan comes back in
// dict_entry order. Measured — category 197 answers ["en-US","fr-FR"] while the
// rows sit in the heap as fr-FR then en-US, so neither id nor ctid explains it.
//
// Ordering explicitly by dict_entry reproduces today's output. Java's own order
// is incidental and a different plan would change it, which makes this a
// D-class item: the array order is observable in the response and nothing in
// the Java code guarantees it.
func (dao *SiteInformationDAOImpl) DictionaryEntriesByCategory(categoryID int64) ([]string, error) {
	values := []string{}
	err := dao.DB.Table("clinlims.dictionary").
		Select("dict_entry").
		Where("dictionary_category_id = ?", categoryID).
		Order("dict_entry").
		Scan(&values).Error
	return values, err
}

// Insert ports the new-row branch of validateAndUpdateSiteInformation.
//
// Only these columns are written. value_type is FORCED to 'text' by the caller
// whatever the request asked for, and the domain is the site-identity default —
// see the service.
func (dao *SiteInformationDAOImpl) Insert(row *valueholder.SiteInformation, sysUserID int64) error {
	return dao.DB.Transaction(func(tx *gorm.DB) error {
		ts := writeTime()
		var id string
		err := tx.Raw(`
			INSERT INTO clinlims.site_information
			       (id, name, description, value, encrypted, domain_id, value_type, "group", lastupdated)
			VALUES (nextval('clinlims.site_information_seq'), ?, ?, ?, ?, ?, ?, 0, ?)
			RETURNING id::text`,
			row.Name, row.Description, row.Value, row.Encrypted, row.DomainID, row.ValueType, ts).
			Scan(&id).Error
		if err != nil {
			return err
		}
		// saveNewHistory sets no payload, so `changes` is NULL rather than empty.
		return dao.Audit.Write(tx, siteInformationTable, id, sysUserID, audittrail.ActivityInsert, nil, ts)
	})
}

// UpdateValue ports the existing-row branch, and the name is the whole point:
// the update writes the VALUE and nothing else.
//
// validateAndUpdateSiteInformation loads the row by id and sets only setValue
// and setSysUserId before persisting, so paramName and description are read off
// the submitted form, echoed back in the response, and never stored. A rename
// through this endpoint reports success and changes nothing.
func (dao *SiteInformationDAOImpl) UpdateValue(id string, value *string, sysUserID int64) error {
	return dao.DB.Transaction(func(tx *gorm.DB) error {
		var old []string
		if err := tx.Raw(
			`SELECT COALESCE(value, '') FROM clinlims.site_information WHERE id = ?`, id).
			Scan(&old).Error; err != nil {
			return err
		}
		if len(old) == 0 {
			return ErrNoSuchRow
		}

		// An update that changes nothing writes NOTHING — not the row, not
		// lastupdated, not an audit entry. Measured: posting a row's own value
		// back left its lastupdated at the value it has carried since 2020 and
		// added no history row. Hibernate finds the entity clean and skips the
		// UPDATE, and getChanges returns empty so saveHistory has nothing to
		// record. A port that always ran the UPDATE would bump lastupdated on
		// every save button press.
		next := ""
		if value != nil {
			next = *value
		}
		if next == old[0] {
			return nil
		}

		ts := writeTime()
		if err := tx.Exec(
			`UPDATE clinlims.site_information SET value = ?, lastupdated = ? WHERE id = ?`,
			value, ts, id).Error; err != nil {
			return err
		}
		// The payload is the value being REPLACED, not the new one.
		changes := audittrail.Field("value", old[0])
		return dao.Audit.Write(tx, siteInformationTable, id, sysUserID, audittrail.ActivityUpdate, &changes, ts)
	})
}

// DeleteAll ports siteInformationService.deleteAll.
func (dao *SiteInformationDAOImpl) DeleteAll(ids []string, sysUserID int64) error {
	if len(ids) == 0 {
		return nil
	}
	return dao.DB.Transaction(func(tx *gorm.DB) error {
		// The payload is built BEFORE the delete, from the row that is about to
		// stop existing.
		type doomed struct {
			ID          string  `gorm:"column:id"`
			Name        string  `gorm:"column:name"`
			Description string  `gorm:"column:description"`
			Value       string  `gorm:"column:value"`
			Encrypted   bool    `gorm:"column:encrypted"`
			ValueType   string  `gorm:"column:value_type"`
			Domain      string  `gorm:"column:domain"`
			Group       int     `gorm:"column:grp"`
			Schedule    string  `gorm:"column:schedule"`
			Lastupdated *string `gorm:"column:lastupdated"`
		}
		rows := []doomed{}
		err := tx.Table("clinlims.site_information AS si").
			Select(`si.id::text AS id, si.name,
				COALESCE(si.description, '') AS description,
				COALESCE(si.value, '') AS value,
				COALESCE(si.encrypted, false) AS encrypted,
				si.value_type,
				COALESCE(d.name, '') AS domain,
				si."group" AS grp,
				'' AS schedule,
				to_char(si.lastupdated AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI:SS.MS') AS lastupdated`).
			Joins("LEFT JOIN clinlims.site_information_domain AS d ON d.id = si.domain_id").
			Where("si.id IN ?", ids).
			Order("si.id").
			Scan(&rows).Error
		if err != nil {
			return err
		}

		if err := tx.Exec(`DELETE FROM clinlims.site_information WHERE id IN ?`, ids).Error; err != nil {
			return err
		}

		ts := writeTime()
		for _, r := range rows {
			// The field list and its ORDER are the wire contract of the audit
			// payload, measured off a delete Java performed.
			//
			// tag, instructionKey, dictionaryCategoryId, descriptionKey and
			// nameKey are absent, and that is not an omission: getChanges
			// compares the row against a BLANK object and emits only the fields
			// that differ, so a NULL column matches the blank and drops out.
			// Every row reachable through this endpoint has them NULL — the
			// insert path never sets them — so the list is fixed here. A row
			// that carries a tag would need the general mechanism; see
			// open-items.md.
			//
			// domain is the domain NAME, not the id. schedule is always empty:
			// no site_information row has a schedule_id.
			changes := audittrail.Field("name", r.Name) +
				audittrail.Field("description", r.Description) +
				audittrail.Field("value", r.Value) +
				audittrail.Field("encrypted", strconv.FormatBool(r.Encrypted)) +
				audittrail.Field("valueType", r.ValueType) +
				audittrail.Field("domain", r.Domain) +
				audittrail.Field("group", strconv.Itoa(r.Group)) +
				audittrail.Field("schedule", r.Schedule) +
				audittrail.Field("lastupdated", derefTime(r.Lastupdated))
			if err := dao.Audit.Write(tx, siteInformationTable, r.ID, sysUserID,
				audittrail.ActivityDelete, &changes, ts); err != nil {
				return err
			}
		}
		return nil
	})
}

func derefTime(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// LocalizationRow is one localization_value joined to its localization.
type LocalizationRow struct {
	LocalizationID          string  `gorm:"column:localization_id"`
	LocalizationDescription *string `gorm:"column:localization_description"`
	LocalizationUpdated     *int64  `gorm:"column:localization_updated"`
	ValueID                 string  `gorm:"column:value_id"`
	Locale                  string  `gorm:"column:locale"`
	Value                   string  `gorm:"column:value"`
	ValueUpdated            *int64  `gorm:"column:value_updated"`
}

// LocalizationByID loads one localization and every one of its values.
//
// The menu handler reaches this only for a row tagged "localization", where
// the site_information VALUE column is not a value at all — it is the
// localization id. createMenuList swaps the whole Localization object in.
func (dao *SiteInformationDAOImpl) LocalizationByID(id string) ([]LocalizationRow, error) {
	rows := []LocalizationRow{}
	err := dao.DB.Table("clinlims.localization AS l").
		Select(`l.id::text AS localization_id,
			l.description AS localization_description,
			trunc(EXTRACT(EPOCH FROM l.lastupdated) * 1000)::bigint AS localization_updated,
			lv.id::text AS value_id,
			lv.locale AS locale,
			COALESCE(lv.value, '') AS value,
			trunc(EXTRACT(EPOCH FROM lv.last_updated) * 1000)::bigint AS value_updated`).
		Joins("JOIN clinlims.localization_value AS lv ON lv.localization_id = l.id").
		Where("l.id = ?", id).
		Order("lv.locale").
		Scan(&rows).Error
	return rows, err
}

// ActiveLocales ports LocalizationService.getAllActiveLocales.
func (dao *SiteInformationDAOImpl) ActiveLocales() ([]string, error) {
	codes := []string{}
	err := dao.DB.Table("clinlims.supported_locale").
		Select("locale_code").
		Where("is_active = ?", true).
		Order("sort_order").
		Scan(&codes).Error
	return codes, err
}
