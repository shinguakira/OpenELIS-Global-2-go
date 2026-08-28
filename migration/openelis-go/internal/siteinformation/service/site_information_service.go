// Package service ports org.openelisglobal.siteinformation.service plus the
// per-request form setup the Java controllers keep inline.
//
// Per constitution.md Layer III, the DTO is assembled here; the controller does
// request/response mapping only.
package service

import (
	"strconv"
	"strings"

	locform "openelis-go/internal/localization/form"
	"openelis-go/internal/siteinformation/daoimpl"
	"openelis-go/internal/siteinformation/form"
	"openelis-go/internal/siteinformation/valueholder"
)

// SiteIdentityDomainID is SITE_IDENTITY_DOMAIN — the domain every row created
// through this endpoint lands in, whatever the caller asked for.
const SiteIdentityDomainID int64 = 1

// ResultConfigDomainID is RESULT_CONFIG_DOMAIN.
const ResultConfigDomainID int64 = 9

// domainRoute is one row of setupFormForRequest's if/else ladder.
//
// A SLICE, not a map: Java dispatches with a chain of path.contains(...) tests
// in a fixed order, and the order is load-bearing. "NextPreviousResultConfiguration"
// contains "ResultConfiguration"; "PrintedReportsConfiguration" is tested before
// "SampleEntryConfig". Iterating a map would pick a different branch on
// different runs.
type domainRoute struct {
	Match string // the substring path.contains() looks for

	// FormDomain is what the FORM controller reports as siteInfoDomainName.
	FormDomain string
	// MenuDomain is what the MENU controller reports for the SAME domain, and
	// for PatientConfiguration the two DISAGREE — the form controller spells it
	// "PaitientConfiguration" and the menu controller spells it correctly. One
	// domain, two names, decided by which endpoint you ask.
	MenuDomain string

	FormName string // form controller's formName
	MenuName string // menu controller's formName
	Action   string // formAction, and "Cancel"+Action is cancelAction

	// DBDomain is the site_information_domain.name the menu filters on. It is a
	// third spelling again: sampleEntryConfig here, printedReportsConfig where
	// the domain label says PrintedReportsConfiguration.
	DBDomain string
}

// domainRoutes is the ladder, in Java's order. The final entry is the `else`
// branch and matches everything, so it must stay last.
var domainRoutes = []domainRoute{
	{"NonConformityConfiguration", "non_conformityConfiguration", "non_conformityConfiguration",
		"NonConformityConfigurationForm", "NonConformityConfigurationMenuForm",
		"NonConformityConfiguration", "non_conformityConfig"},
	{"WorkplanConfiguration", "WorkplanConfiguration", "WorkplanConfiguration",
		"WorkplanConfigurationForm", "WorkplanConfigurationMenuForm",
		"WorkplanConfiguration", "workplanConfig"},
	{"PrintedReportsConfiguration", "PrintedReportsConfiguration", "PrintedReportsConfiguration",
		"PrintedReportsConfigurationForm", "PrintedReportsConfigurationMenuForm",
		"PrintedReportsConfiguration", "printedReportsConfig"},
	{"SampleEntryConfig", "sampleEntryConfig", "sampleEntryConfig",
		"sampleEntryConfigForm", "sampleEntryConfigMenuForm",
		"SampleEntryConfig", "sampleEntryConfig"},
	{"ResultConfiguration", "ResultConfiguration", "ResultConfiguration",
		"resultConfigurationForm", "resultConfigurationMenuForm",
		"ResultConfiguration", "resultConfiguration"},
	{"MenuStatementConfig", "MenuStatementConfig", "MenuStatementConfig",
		"MenuStatementConfigForm", "MenuStatementConfigMenuForm",
		"MenuStatementConfig", "MenuStatementConfig"},
	{"PatientConfiguration", "PaitientConfiguration", "PatientConfiguration",
		"PatientConfigurationForm", "PatientConfigurationMenuForm",
		"PatientConfiguration", "patientEntryConfig"},
	{"ValidationConfiguration", "validationConfig", "validationConfig",
		"ValidationConfigurationForm", "ValidationConfigurationMenuForm",
		"ValidationConfiguration", "validationConfig"},
	{"", "SiteInformation", "SiteInformation",
		"SiteInformationForm", "siteInformationMenuForm",
		"SiteInformation", "siteIdentity"},
}

// RouteFor is setupFormForRequest's dispatch, on the servlet path.
func RouteFor(path string) domainRoute {
	for _, r := range domainRoutes {
		if r.Match == "" || strings.Contains(path, r.Match) {
			return r
		}
	}
	return domainRoutes[len(domainRoutes)-1]
}

// SiteInformationService ports SiteInformationServiceImpl.
type SiteInformationService struct {
	DAO *daoimpl.SiteInformationDAOImpl
	// Msgs is the parsed message bundle, used to resolve instructionKey and
	// descriptionKey into the `description` the form shows.
	Msgs map[string]string
	// ActiveLocale is site_information "default language locale", trimmed to
	// its language part.
	ActiveLocale string
}

// Show ports showSiteInformation.
//
// The id is read from the `ID` query parameter — UPPERCASE. `?id=83` is not a
// synonym: it misses, isNew stays true, and the endpoint answers the blank
// "add new" form for a row that exists. Reproduced deliberately; the caller
// passes whatever it read from `ID` and nothing else.
func (s *SiteInformationService) Show(path, id string) (*form.SiteInformationForm, bool, error) {
	f := form.NewSiteInformationForm()
	route := RouteFor(path)
	f.FormName = route.FormName
	f.FormAction = route.Action
	f.CancelAction = "Cancel" + route.Action
	f.SiteInfoDomainName = route.FormDomain

	if id == "" || id == "0" {
		return &f, true, nil
	}

	row, err := s.DAO.Get(id)
	if err != nil {
		return nil, false, err
	}
	if row == nil {
		// Java dereferences the null here and the request ends as a 500. The
		// controller turns this false into that same 500 rather than a 404.
		return nil, false, nil
	}

	f.ParamName = row.Name
	f.Description = s.instruction(row)
	f.Value = derefStr(row.Value)
	f.Encrypted = row.Encrypted != nil && *row.Encrypted
	f.ValueType = row.ValueType
	f.Editable = isEditable(row.Name)
	f.Tag = row.Tag

	if row.ValueType == "dictionary" && row.DictionaryCategoryID != nil {
		values, err := s.DAO.DictionaryEntriesByCategory(*row.DictionaryCategoryID)
		if err != nil {
			return nil, false, err
		}
		f.DictionaryValues = values
	}
	return &f, true, nil
}

// accessionNumberPrefix is the controller's ACCESSION_NUMBER_PREFIX constant.
const accessionNumberPrefix = "Accession number prefix"

// isEditable ports isEditable, suffix test and all.
//
// Java asks whether the CONSTANT ends with the row's NAME —
// `ACCESSION_NUMBER_PREFIX.endsWith(siteInformation.getName())` — where
// equality was plainly meant. Any row named "prefix", "number prefix" or even
// "x" would be treated as the accession-number prefix and locked once samples
// exist. No such row is shipped, so the two readings agree on this data.
//
// The sample-count half is not ported: it returns false only when the accession
// prefix row is edited on a database with zero samples, and this deployment has
// samples. Left as a note rather than a silent simplification.
func isEditable(name string) bool {
	return !strings.HasSuffix(accessionNumberPrefix, name)
}

// instruction ports getInstruction: the instruction message, else the
// description message, else the raw description column — first non-blank wins.
func (s *SiteInformationService) instruction(row *valueholder.SiteInformation) string {
	if row.InstructionKey != nil {
		if v := s.Msgs[*row.InstructionKey]; v != "" {
			return v
		}
	}
	if row.DescriptionKey != nil {
		if v := s.Msgs[*row.DescriptionKey]; v != "" {
			return v
		}
	}
	return derefStr(row.Description)
}

// Menu ports showSiteInformationMenu + createMenuList.
func (s *SiteInformationService) Menu(path string) (*form.SiteInformationMenuForm, error) {
	route := RouteFor(path)
	rows, err := s.DAO.ByDomainName(route.DBDomain)
	if err != nil {
		return nil, err
	}

	// The active locale list is needed only when a row carries a localization,
	// so it is loaded once and lazily — most domains have none.
	var activeLocales []string
	items := make([]form.MenuItem, 0, len(rows))
	for i := range rows {
		item := toMenuItem(&rows[i])
		if derefStr(rows[i].Tag) == "localization" {
			if activeLocales == nil {
				activeLocales, err = s.DAO.ActiveLocales()
				if err != nil {
					return nil, err
				}
			}
			loc, err := s.localization(derefStr(rows[i].Value), activeLocales)
			if err != nil {
				return nil, err
			}
			item.Localization = loc
		}
		items = append(items, item)
	}

	return &form.SiteInformationMenuForm{
		FormName:       route.MenuName,
		FormMethod:     "POST",
		CancelAction:   "Home",
		SubmitOnCancel: false,
		CancelMethod:   "POST",
		// Constant, on every domain and every size — see the form type.
		TotalRecordCount:   "",
		FromRecordCount:    "1",
		ToRecordCount:      "20",
		MenuList:           items,
		SelectedIDs:        []string{},
		SiteInfoDomainName: route.MenuDomain,
	}, nil
}

func toMenuItem(row *valueholder.SiteInformation) form.MenuItem {
	item := form.MenuItem{
		ID:             strconv.FormatInt(row.ID, 10),
		Name:           row.Name,
		Description:    row.Description,
		Value:          row.Value,
		Encrypted:      row.Encrypted != nil && *row.Encrypted,
		ValueType:      row.ValueType,
		InstructionKey: row.InstructionKey,
		Group:          row.Group,
		Tag:            row.Tag,
		DescriptionKey: row.DescriptionKey,
	}
	if row.DictionaryCategoryID != nil {
		id := strconv.FormatInt(*row.DictionaryCategoryID, 10)
		item.DictionaryCategoryID = &id
	}
	if row.Lastupdated != nil {
		ms := row.Lastupdated.UnixMilli()
		item.Lastupdated = &ms
	}
	if row.DomainID != nil {
		item.Domain = &form.DomainDTO{
			ID:          strconv.FormatInt(*row.DomainID, 10),
			Name:        derefStr(row.DomainName),
			Description: row.DomainDescription,
		}
	}
	// hideEncryptedFields: an encrypted row's value is replaced character for
	// character with '*'. Java does it with value.replaceAll(".", "*") — a
	// REGEX whose dot matches any character, so the mask is the same length as
	// the value.
	//
	// KNOWN GAP, stated rather than hidden. Java masks the DECRYPTED plaintext:
	// SiteInformationServiceImpl decrypts through a jasypt AES256TextEncryptor
	// on every read, so a row holding "secret-value" is stored as 64 base64
	// characters and comes back as TWELVE asterisks. This masks the stored
	// column, so it would answer sixty-four.
	//
	// Measured — a row created through Java's own POST with encrypted=true
	// stored 48 bytes (16-byte salt, 16-byte IV, one AES block) and the menu
	// rendered 12 asterisks. Reproducing it needs the encryptor ported, and the
	// exact key derivation is not yet pinned: the password is the documented
	// default "dev" and the layout is three 16-byte groups, but PBKDF2-HMAC
	// over SHA-512/256/1 at 1..5000 iterations does not reproduce the
	// ciphertext, so the parameters have to be read out of jasypt rather than
	// guessed at.
	//
	// NOT reachable on stock data: no shipped row carries encrypted = true, and
	// the fixture that would have created one is deliberately absent — a
	// site_information row with a plaintext value in that column makes Java 500
	// on every config menu, because the decrypt throws and the resulting error
	// object is itself unserialisable. Seeding it wrong is worse than not
	// seeding it.
	if item.Encrypted && item.Value != nil && *item.Value != "" {
		masked := strings.Repeat("*", len([]rune(*item.Value)))
		item.Value = &masked
	}
	return item
}

// Update ports validateAndUpdateSiteInformation.
//
// Two branches, and both of them store less than the response implies:
//
//   - NEW: name, description, encrypted and the domain are taken from the form,
//     but value_type is FORCED to "text" — `setValueType("text")` — while the
//     response echoes back whatever valueType was submitted. Send "boolean",
//     store "text", get told "boolean".
//   - EXISTING: the row is loaded by id and ONLY setValue is called. paramName
//     and description are read off the form, echoed back, and never persisted.
//
// The returned form is the REQUEST BODY over the bean defaults, because the
// POST path never calls setupFormForRequest.
func (s *SiteInformationService) Update(post form.SiteInformationPost, id string) (*form.SiteInformationForm, error) {
	isNew := id == "" || id == "0"
	name := derefStr(post.ParamName)
	value := post.Value

	if isNew {
		domainID := SiteIdentityDomainID
		if derefStr(post.SiteInfoDomainName) == "ResultConfiguration" {
			domainID = ResultConfigDomainID
		}
		encrypted := post.Encrypted != nil && *post.Encrypted
		row := &valueholder.SiteInformation{
			Name:        name,
			Description: post.Description,
			Value:       value,
			Encrypted:   &encrypted,
			DomainID:    &domainID,
			ValueType:   "text",
		}
		if err := s.DAO.Insert(row); err != nil {
			return nil, err
		}
	} else {
		if err := s.DAO.UpdateValue(id, value); err != nil {
			return nil, err
		}
	}

	return echoForm(post), nil
}

// echoForm is the POST response: the deserialised body applied over
// `new SiteInformationForm()`. Absent keys keep the bean default, which is why
// a body with no tag still answers `"tag":""`.
func echoForm(post form.SiteInformationPost) *form.SiteInformationForm {
	f := form.NewSiteInformationForm()
	if post.ParamName != nil {
		f.ParamName = *post.ParamName
	}
	if post.Description != nil {
		f.Description = *post.Description
	}
	if post.Value != nil {
		f.Value = *post.Value
	}
	if post.Encrypted != nil {
		f.Encrypted = *post.Encrypted
	}
	if post.ValueType != nil {
		f.ValueType = *post.ValueType
	}
	if post.SiteInfoDomainName != nil {
		f.SiteInfoDomainName = *post.SiteInfoDomainName
	}
	if post.Editable != nil {
		f.Editable = *post.Editable
	}
	if post.Tag != nil {
		f.Tag = post.Tag
	}
	if post.DescriptionKey != nil {
		f.DescriptionKey = *post.DescriptionKey
	}
	return &f
}

// Delete ports showDeleteSiteInformation's happy path.
func (s *SiteInformationService) Delete(ids []string) error {
	return s.DAO.DeleteAll(ids)
}

// localization loads and assembles the nested Localization object.
//
// requestLocale is the deployment locale — LocaleContextHolder.getLocale() in
// Java. localizedValue and the display-name pairs are resolved against it.
func (s *SiteInformationService) localization(localizationID string, activeLocales []string) (*locform.LocalizationDTO, error) {
	rows, err := s.DAO.LocalizationByID(localizationID)
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	values := make([]locform.LocalizationValue, 0, len(rows))
	for _, r := range rows {
		values = append(values, locform.LocalizationValue{
			Lastupdated: r.ValueUpdated,
			ID:          r.ValueID,
			Locale:      r.Locale,
			Value:       r.Value,
		})
	}
	return locform.BuildLocalization(rows[0].LocalizationID, derefStr(rows[0].LocalizationDescription),
		rows[0].LocalizationUpdated, values, activeLocales, s.Locale()), nil
}

// Locale is the configured request locale, defaulting to English.
func (s *SiteInformationService) Locale() string {
	if s.ActiveLocale != "" {
		return s.ActiveLocale
	}
	return "en"
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
