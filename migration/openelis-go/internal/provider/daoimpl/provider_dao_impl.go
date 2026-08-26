// Package daoimpl ports org.openelisglobal.provider.daoimpl +
// org.openelisglobal.person.daoimpl (the methods the ported endpoints use).
// Folder layout mirrors the Java source during migration.
package daoimpl

import (
	"errors"
	"fmt"
	"regexp"

	"gorm.io/gorm"

	"openelis-go/internal/provider/valueholder"
)

// ProviderDAOImpl ports ProviderDAOImpl + PersonDAOImpl + the generic
// BaseDAOImpl paths the ported (read-only) endpoints actually use.
type ProviderDAOImpl struct {
	DB *gorm.DB
}

// GetProviderByID mirrors the generic BaseObjectServiceImpl.get(id) path
// used by GET Provider/raw/{id}: entityManager.find(Provider.class, id).
// Returns (nil, nil) on no match — see the doc comment on
// organization/daoimpl.OrganizationDAOImpl.GetByID for why this port
// returns 404 rather than reproducing Java's confirmed 500-on-not-found bug
// (same root cause here: Provider/raw/{id}'s Java controller has the same
// dead not-found branch).
func (d *ProviderDAOImpl) GetProviderByID(id int64) (*valueholder.Provider, error) {
	var p valueholder.Provider
	result := d.DB.First(&p, "id = ?", id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &p, nil
}

// GetPersonByID mirrors both PersonDAOImpl.getPersonById (used by GET
// rest/practitioner) and the generic BaseObjectServiceImpl.get(id) path
// (used by GET Provider/Person/{id}) — both resolve to the same
// SELECT * FROM person WHERE id = ? in practice. Returns (nil, nil) on no
// match — same 404-not-500 divergence as GetProviderByID.
func (d *ProviderDAOImpl) GetPersonByID(id int64) (*valueholder.Person, error) {
	var p valueholder.Person
	result := d.DB.First(&p, "id = ?", id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &p, nil
}

// GetProviderByPersonID mirrors ProviderDAOImpl.getProviderByPerson: HQL
// "from Provider p where p.person.id = :personId", first row if more than
// one (the DAO's own comment documents the app's 1:1-in-practice
// assumption; the DB doesn't enforce uniqueness on person_id). Returns
// (nil, nil) when no provider is linked to that person.
func (d *ProviderDAOImpl) GetProviderByPersonID(personID int64) (*valueholder.Provider, error) {
	var providers []valueholder.Provider
	if err := d.DB.Where("person_id = ?", personID).Order("id ASC").Limit(1).Find(&providers).Error; err != nil {
		return nil, err
	}
	if len(providers) == 0 {
		return nil, nil
	}
	return &providers[0], nil
}

// ProviderSearchRow is the scan target for the provider<->person JOIN used
// by all three GetPage/SearchByName/SearchByPhone variants below.
type ProviderSearchRow struct {
	ProviderID   int64
	PersonID     int64
	FirstName    *string
	LastName     *string
	PrimaryPhone *string
	WorkPhone    *string
	Fax          *string
	Email        *string
	ExternalID   *string
	Active       *bool
}

const providerSearchSelect = `p.id AS provider_id, p.person_id AS person_id,
	pe.first_name AS first_name, pe.last_name AS last_name,
	pe.primary_phone AS primary_phone, pe.work_phone AS work_phone,
	pe.fax AS fax, pe.email AS email,
	p.external_id AS external_id, p.active AS active`

func (d *ProviderDAOImpl) providerPersonJoin() *gorm.DB {
	return d.DB.Table("clinlims.provider AS p").
		Joins("JOIN clinlims.person pe ON pe.id = p.person_id")
}

// GetPage mirrors ProviderDAOImpl.getPageOfProviders: "from Provider p
// order by p.id" — no filter, no active-only restriction, plain id order.
// Used by GET provider/search when neither ?search nor ?phone is given.
func (d *ProviderDAOImpl) GetPage(offset, limit int) ([]ProviderSearchRow, error) {
	var rows []ProviderSearchRow
	err := d.providerPersonJoin().Select(providerSearchSelect).
		Order("p.id ASC").Offset(offset).Limit(limit).Find(&rows).Error
	return rows, err
}

// Count mirrors the generic BaseDAOImpl.getCount() path used as the
// fallback totalCount when GetPage is used.
func (d *ProviderDAOImpl) Count() (int64, error) {
	var count int64
	err := d.DB.Table("clinlims.provider").Count(&count).Error
	return count, err
}

// SearchByName mirrors ProviderDAOImpl.getPagesOfSearchedProviders: matches
// first name, last name, or "first last" concatenation, case-insensitive
// substring — HQL's `lower(p.person.firstName) like concat('%',
// lower(:searchValue), '%')` translated directly to SQL LIKE. Sorted
// active-first, then by last name — matches the real ORDER BY exactly.
func (d *ProviderDAOImpl) SearchByName(term string, offset, limit int) ([]ProviderSearchRow, error) {
	var rows []ProviderSearchRow
	needle := "%" + term + "%"
	err := d.providerPersonJoin().Select(providerSearchSelect).
		Where("lower(pe.first_name) LIKE lower(?) OR lower(pe.last_name) LIKE lower(?) OR lower(pe.first_name || ' ' || pe.last_name) LIKE lower(?)", needle, needle, needle).
		Order("p.active DESC, pe.last_name ASC").
		Offset(offset).Limit(limit).Find(&rows).Error
	return rows, err
}

// CountByName mirrors ProviderDAOImpl.getTotalSearchedProviderCount — same
// WHERE as SearchByName, COUNT(*) instead of a page of rows.
func (d *ProviderDAOImpl) CountByName(term string) (int64, error) {
	var count int64
	needle := "%" + term + "%"
	err := d.providerPersonJoin().
		Where("lower(pe.first_name) LIKE lower(?) OR lower(pe.last_name) LIKE lower(?) OR lower(pe.first_name || ' ' || pe.last_name) LIKE lower(?)", needle, needle, needle).
		Count(&count).Error
	return count, err
}

var nonDigitRE = regexp.MustCompile(`[^0-9]`)

// digitsOnly mirrors ProviderDAOImpl's phone-search normalization:
// phone.replaceAll("[^0-9]", "").
func digitsOnly(phone string) string {
	return nonDigitRE.ReplaceAllString(phone, "")
}

const stripPhonePunctuationSQL = `replace(replace(replace(replace(%s, ' ', ''), '-', ''), '(', ''), ')', '')`

// SearchByPhone mirrors ProviderDAOImpl.getPagesOfSearchedProvidersByPhone:
// strips spaces/-/()  from primary_phone, work_phone, and fax, then does a
// substring match against the caller's digits-only input. Empty
// digits-only input returns no rows (matches Java's early-return []).
func (d *ProviderDAOImpl) SearchByPhone(phone string, offset, limit int) ([]ProviderSearchRow, error) {
	digits := digitsOnly(phone)
	if digits == "" {
		return []ProviderSearchRow{}, nil
	}
	needle := "%" + digits + "%"
	primary := fmt.Sprintf(stripPhonePunctuationSQL, "pe.primary_phone")
	work := fmt.Sprintf(stripPhonePunctuationSQL, "pe.work_phone")
	fax := fmt.Sprintf(stripPhonePunctuationSQL, "pe.fax")
	var rows []ProviderSearchRow
	err := d.providerPersonJoin().Select(providerSearchSelect).
		Where(primary+" LIKE ? OR "+work+" LIKE ? OR "+fax+" LIKE ?", needle, needle, needle).
		Order("p.active DESC, pe.last_name ASC").
		Offset(offset).Limit(limit).Find(&rows).Error
	return rows, err
}

// CountByPhone mirrors ProviderDAOImpl.getTotalSearchedProviderCountByPhone.
func (d *ProviderDAOImpl) CountByPhone(phone string) (int64, error) {
	digits := digitsOnly(phone)
	if digits == "" {
		return 0, nil
	}
	needle := "%" + digits + "%"
	primary := fmt.Sprintf(stripPhonePunctuationSQL, "pe.primary_phone")
	work := fmt.Sprintf(stripPhonePunctuationSQL, "pe.work_phone")
	fax := fmt.Sprintf(stripPhonePunctuationSQL, "pe.fax")
	var count int64
	err := d.providerPersonJoin().
		Where(primary+" LIKE ? OR "+work+" LIKE ? OR "+fax+" LIKE ?", needle, needle, needle).
		Count(&count).Error
	return count, err
}
