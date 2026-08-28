// Package daoimpl ports org.openelisglobal.siteinformation.daoimpl.
// Folder layout mirrors the Java source during migration.
package daoimpl

import (
	"errors"

	"gorm.io/gorm"

	"openelis-go/internal/siteinformation/valueholder"
)

// SiteInformationDAOImpl ports SiteInformationDAOImpl.
type SiteInformationDAOImpl struct {
	DB *gorm.DB
}

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
func (dao *SiteInformationDAOImpl) Insert(row *valueholder.SiteInformation) error {
	return dao.DB.Exec(`
		INSERT INTO clinlims.site_information
		       (id, name, description, value, encrypted, domain_id, value_type, "group", lastupdated)
		VALUES (nextval('clinlims.site_information_seq'), ?, ?, ?, ?, ?, ?, 0, now())`,
		row.Name, row.Description, row.Value, row.Encrypted, row.DomainID, row.ValueType).Error
}

// UpdateValue ports the existing-row branch, and the name is the whole point:
// the update writes the VALUE and nothing else.
//
// validateAndUpdateSiteInformation loads the row by id and sets only setValue
// and setSysUserId before persisting, so paramName and description are read off
// the submitted form, echoed back in the response, and never stored. A rename
// through this endpoint reports success and changes nothing.
func (dao *SiteInformationDAOImpl) UpdateValue(id string, value *string) error {
	return dao.DB.Exec(
		`UPDATE clinlims.site_information SET value = ?, lastupdated = now() WHERE id = ?`,
		value, id).Error
}

// DeleteAll ports siteInformationService.deleteAll.
func (dao *SiteInformationDAOImpl) DeleteAll(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return dao.DB.Exec(`DELETE FROM clinlims.site_information WHERE id IN ?`, ids).Error
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
