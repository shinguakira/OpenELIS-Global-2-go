// Package daoimpl ports org.openelisglobal.localization.daoimpl.
// Folder layout mirrors the Java source during migration.
package daoimpl

import (
	"errors"

	"gorm.io/gorm"

	"openelis-go/internal/localization/valueholder"
)

// SupportedLocaleDAO ports SupportedLocaleDAOImpl — reads clinlims.supported_locale
// via GORM's query builder. No .Select() needed: every column maps 1:1 to a
// struct field via gorm tags, and none are nullable (no COALESCE needed, same
// as the original Java query).
type SupportedLocaleDAO struct {
	DB *gorm.DB
}

// GetAll mirrors BaseDAOImpl.getAll() — every row, no ORDER BY (DB-natural order).
func (d *SupportedLocaleDAO) GetAll() ([]valueholder.SupportedLocale, error) {
	var list []valueholder.SupportedLocale
	result := d.DB.Find(&list)
	if list == nil {
		list = []valueholder.SupportedLocale{}
	}
	return list, result.Error
}

// GetAllActive mirrors getAllMatchingOrdered("active", true, "sortOrder", false):
// WHERE is_active = true ORDER BY sort_order ASC.
func (d *SupportedLocaleDAO) GetAllActive() ([]valueholder.SupportedLocale, error) {
	var list []valueholder.SupportedLocale
	result := d.DB.Where("is_active = ?", true).Order("sort_order ASC").Find(&list)
	if list == nil {
		list = []valueholder.SupportedLocale{}
	}
	return list, result.Error
}

// GetFallback mirrors getFallback() — the first row with is_fallback = true, or
// nil when none (the controller turns nil into 404). Uses First() (GORM's
// canonical single-row read, mirrors JPA's getSingleResult()/NoResultException)
// instead of Find()+manual empty check.
func (d *SupportedLocaleDAO) GetFallback() (*valueholder.SupportedLocale, error) {
	var loc valueholder.SupportedLocale
	err := d.DB.Where("is_fallback = ?", true).First(&loc).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &loc, nil
}
