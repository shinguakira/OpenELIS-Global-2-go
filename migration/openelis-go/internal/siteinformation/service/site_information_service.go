// Package service ports org.openelisglobal.siteinformation.service plus the
// per-request form setup the Java controllers keep inline.
//
// Per constitution.md Layer III, the DTO is assembled here; the controller does
// request/response mapping only.
package service

import (
	"regexp"
	"strconv"
	"strings"
	"sync"

	"gorm.io/gorm"

	locform "openelis-go/internal/localization/form"
	"openelis-go/internal/security/encryption"
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
	// Encryptor is the jasypt TextEncryptor bean. A row flagged `encrypted`
	// stores ciphertext, and the service decrypts on read and encrypts on
	// write — so the value the form shows and the value the column holds are
	// never the same string.
	Encryptor *encryption.TextEncryptor

	// config is the ConfigurationProperties cache — see Reload.
	configMu sync.RWMutex
	config   map[string]string
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
	f.Value = s.plaintext(row)
	f.Encrypted = row.Encrypted != nil && *row.Encrypted
	f.ValueType = row.ValueType
	f.Tag = row.Tag

	editable, err := s.isEditable(row.Name)
	if err != nil {
		return nil, false, err
	}
	f.Editable = editable

	// setLocalizationValues: a localization-tagged row carries the whole
	// Localization graph, and its `value` column is the localization id.
	if derefStr(row.Tag) == "localization" {
		locales, err := s.DAO.ActiveLocales()
		if err != nil {
			return nil, false, err
		}
		loc, err := s.localization(derefStr(row.Value), locales)
		if err != nil {
			return nil, false, err
		}
		f.Localization = loc
	}

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
// When the suffix test DOES match, the row is editable only while the database
// holds no samples: the accession prefix cannot be changed once numbers have
// been issued under it. This deployment has samples, so the answer is false —
// but it is computed rather than assumed, because a port that returned a
// constant would be right here and wrong on an empty instance.
func (s *SiteInformationService) isEditable(name string) (bool, error) {
	if !strings.HasSuffix(accessionNumberPrefix, name) {
		return true, nil
	}
	n, err := s.DAO.SampleCount()
	if err != nil {
		return false, err
	}
	return n == 0, nil
}

// plaintext is decryptSiteInformation: an encrypted row's column holds
// ciphertext and every read decrypts it.
//
// A value that will not decrypt is returned unchanged rather than failing the
// request. Java does fail — the jasypt exception propagates, the error object
// Spring builds around it is itself unserialisable, and the whole config menu
// answers 500. That is reproduced nowhere on purpose: it needs a row whose
// column was written by something other than this application, and creating one
// takes the screen down for every domain.
func (s *SiteInformationService) plaintext(row *valueholder.SiteInformation) string {
	value := derefStr(row.Value)
	if row.Encrypted == nil || !*row.Encrypted || value == "" || s.Encryptor == nil {
		return value
	}
	plain, err := s.Encryptor.Decrypt(value)
	if err != nil {
		return value
	}
	return plain
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

		// hideEncryptedFields, and it runs on the DECRYPTED value: the service
		// decrypts every row it loads, so the mask is as long as the secret and
		// not as long as the ciphertext. Java does it with
		// value.replaceAll(".", "*") — a REGEX whose dot matches any character.
		//
		// Masking the stored column instead would answer sixty-four asterisks
		// for a twelve-character secret. That is exactly what this port did
		// until the spec asserted the length.
		if item.Encrypted {
			plain := s.plaintext(&rows[i])
			if plain != "" {
				masked := strings.Repeat("*", len([]rune(plain)))
				item.Value = &masked
			}
		}
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
func (s *SiteInformationService) Update(post form.SiteInformationPost, id string, sysUserID int64) (*form.SiteInformationForm, error) {
	echo := echoForm(post)

	// SiteInformationFormValidator. A rejection is NOT an error status: Java
	// calls saveErrors and returns the form, so the response is a 200 carrying
	// the submitted values back — and nothing is written. Measured on all three
	// lists; the only observable difference between accept and reject is
	// whether the row appears.
	if !validForm(post) {
		return echo, nil
	}

	// The localization branch, and it is not a variation on the write below —
	// it writes to a DIFFERENT TABLE. For a row tagged "localization" the
	// site_information `value` column holds a localization id, the real content
	// lives in localization_value, and site_information is not touched at all.
	if derefStr(post.Tag) == "localization" {
		if err := s.updateLocalization(derefStr(post.Value), post.Localization); err != nil {
			return nil, err
		}
		return echo, nil
	}

	isNew := id == "" || id == "0"
	name := derefStr(post.ParamName)
	value := post.Value

	// isValid: the name is required, and four rows carry format rules keyed on
	// that name rather than on any column. Same shape as the validator above —
	// a failure is a 200 with no write.
	if !isValidWrite(name, derefStr(value)) {
		return echo, nil
	}

	// WHICH flag decides encryption depends on the branch, and the two do not
	// agree.
	//
	// validateAndUpdateSiteInformation builds a fresh entity for a new row and
	// calls setEncrypted(form.isEncrypted()) on it — the submission decides. For
	// an existing row it does not build anything: it LOADS the entity by id and
	// calls only setValue, so the flag on the object is the one the column
	// already holds and the submitted one is never read. encryptSiteInformation
	// then tests that object's flag.
	//
	// Reading the submission on both branches — which this port did — breaks
	// the row both ways. Update an encrypted row with `encrypted` omitted (the
	// bean default is false) and the column takes PLAINTEXT while its own flag
	// still says encrypted, so every later read tries to decrypt it and the
	// whole config screen answers 500. Send true for an unencrypted row and the
	// column takes ciphertext nothing will ever decrypt, because no reader looks
	// at a row that is not flagged.
	encrypted := post.Encrypted != nil && *post.Encrypted
	if !isNew {
		stored, err := s.DAO.Get(id)
		if err != nil {
			return nil, err
		}
		if stored == nil {
			// Java dereferences the null the loader returns and the request
			// ends as a 500. The DAO's own error carries that.
			return nil, daoimpl.ErrNoSuchRow
		}
		encrypted = stored.Encrypted != nil && *stored.Encrypted
	}

	// A row flagged encrypted stores CIPHERTEXT. encryptSiteInformation runs
	// before both the insert and the update, so the column never holds the
	// value the caller sent.
	if s.Encryptor != nil && encrypted && value != nil {
		enc, err := s.Encryptor.Encrypt(*value)
		if err != nil {
			return nil, err
		}
		value = &enc
	}

	// persistData is @Transactional, and it covers BOTH the write and
	// configurationSideEffects.siteInformationChanged. They commit together or
	// not at all: a side effect that fails must not leave the configuration row
	// changed, its audit row written, and the dependent role or accession prefix
	// untouched. Running them in separate transactions — which this port did —
	// answers 500 over a half-applied change.
	//
	// ConfigurationProperties.loadDBValuesIntoConfiguration() is NOT in here.
	// Java calls it from the controller after persistData returns, so it sees
	// only committed state.
	err := s.DAO.Tx(func(tx *gorm.DB) error {
		if isNew {
			domainID := SiteIdentityDomainID
			if derefStr(post.SiteInfoDomainName) == "ResultConfiguration" {
				domainID = ResultConfigDomainID
			}
			enc := encrypted
			row := &valueholder.SiteInformation{
				Name:        name,
				Description: post.Description,
				Value:       value,
				Encrypted:   &enc,
				DomainID:    &domainID,
				ValueType:   "text",
			}
			if err := s.DAO.InsertTx(tx, row, sysUserID); err != nil {
				return err
			}
		} else {
			if err := s.DAO.UpdateValueTx(tx, id, value, sysUserID); err != nil {
				return err
			}
		}

		// The value the side effects see is the one on the ENTITY, which by now
		// has been through encryptSiteInformation — so for an encrypted row it
		// is the ciphertext, not what the caller typed. No shipped row is both
		// encrypted and named by a side effect, but the port follows the object
		// rather than the submission because Java does.
		return s.sideEffects(tx, name, derefStr(value), sysUserID)
	})
	if err != nil {
		return nil, err
	}

	// ConfigurationProperties.loadDBValuesIntoConfiguration() — the write makes
	// its own change visible, and nothing else does.
	if err := s.Reload(); err != nil {
		return nil, err
	}
	return echo, nil
}

// validForm ports SiteInformationFormValidator: three allow-lists, and the
// domain list is a THIRD spelling of the same set — it carries
// "externalConnections" and "validationConfig", neither of which the DELETE
// validator accepts, and it carries the misspelled "PaitientConfiguration"
// that the menu controller does not hand out.
func validForm(post form.SiteInformationPost) bool {
	valueTypes := map[string]bool{
		"boolean": true, "logoUpload": true, "text": true,
		"freeText": true, "dictionary": true, "complex": true,
	}
	domains := map[string]bool{
		"externalConnections": true, "non_conformityConfiguration": true,
		"WorkplanConfiguration": true, "PrintedReportsConfiguration": true,
		"sampleEntryConfig": true, "ResultConfiguration": true,
		"MenuStatementConfig": true, "PaitientConfiguration": true,
		"validationConfig": true, "SiteInformation": true,
	}
	// The tag list includes both "" and null, so an absent tag is fine.
	tags := map[string]bool{
		"enable": true, "url": true, "numericOnly": true,
		"programConfiguration": true, "localization": true, "": true,
	}

	valueType := "text" // the bean initialiser, applied when the key is absent
	if post.ValueType != nil {
		valueType = *post.ValueType
	}
	if !valueTypes[valueType] {
		return false
	}
	if !domains[derefStr(post.SiteInfoDomainName)] {
		return false
	}
	return post.Tag == nil || tags[*post.Tag]
}

// phoneFormat is PhoneNumberService.FORMAT_REGEX.
var phoneFormat = regexp.MustCompile(`^[a-zA-Z0-9+()\s\-/|]+$`)

// isValidWrite ports isValid — a required name, plus four rules that key on the
// row's NAME rather than on any column, so the same value is accepted or
// refused depending on which row it is being written to.
func isValidWrite(name, value string) bool {
	if strings.TrimSpace(name) == "" {
		return false
	}
	switch name {
	case "phone format":
		return phoneFormat.MatchString(value)
	case "phone format label", "phone international format label":
		// Blank is allowed here and NOT on "phone format" above.
		return value == "" || phoneFormat.MatchString(value)
	case "phone international validation":
		v := strings.ToUpper(strings.TrimSpace(value))
		return v == "" || v == "NONE" || v == "E164"
	}
	return true
}

// updateLocalization ports validateAndUpdateLocalization.
//
// languageChanged compares only the locales the SUBMITTED object carries a
// value for; when any of them differs, EVERY active locale is rewritten from
// the submission, with English as the fallback for a locale the submission
// omits. So a partial submission can overwrite a locale it never mentioned.
func (s *SiteInformationService) updateLocalization(localizationID string, submitted *locform.LocalizationDTO) error {
	if localizationID == "" || submitted == nil {
		return nil
	}
	stored, err := s.DAO.LocalizationValuesFor(localizationID)
	if err != nil {
		return err
	}
	if len(stored) == 0 {
		return nil
	}

	incoming := func(locale string) string {
		if v, ok := submitted.Values[locale]; ok && v.Value != "" {
			return v.Value
		}
		if v, ok := submitted.Values["en"]; ok && v.Value != "" {
			return v.Value
		}
		return ""
	}

	changed := false
	for locale, v := range submitted.Values {
		if v.Value == "" {
			continue // getLocalesWithValue skips blanks
		}
		if v.Value != stored[locale] {
			changed = true
		}
	}
	if !changed {
		return nil
	}

	active, err := s.DAO.ActiveLocales()
	if err != nil {
		return err
	}
	values := map[string]string{}
	for _, locale := range active {
		values[locale] = incoming(locale)
	}
	return s.DAO.UpdateLocalizationValues(localizationID, values)
}

// sideEffects ports ConfigurationSideEffects.siteInformationChanged: writing
// one configuration row can change a different table.
// It runs inside the caller's transaction — see Update — because Java's
// persistData wraps the primary write and this together.
func (s *SiteInformationService) sideEffects(tx *gorm.DB, name, value string, sysUserID int64) error {
	// The names are the PROPERTY DB names, not the enum constants: the Java code
	// reads Property.roleRequiredForModifyResults.getDBName(), which is the
	// string "modify results role", and Property.SiteCode.getDBName(), which is
	// "siteNumber". A port that matched on the constant names would compile,
	// read plausibly, and never fire.
	switch name {
	case "modify results role":
		// The setting IS the role's active flag. A port that stored the row and
		// stopped would leave the permission out of step with the screen.
		return s.DAO.SetRoleActiveTx(tx, "Results modifier", value == "true")

	case "siteNumber":
		// Only when the accession format is SITEYEARNUM, and only to FILL a
		// blank prefix — an existing prefix is left alone, because accession
		// numbers have already been issued under it.
		format, err := s.DAO.ByNameTx(tx, "acessionFormat") // misspelled in the data
		if err != nil || format == nil || derefStr(format.Value) != "SITEYEARNUM" {
			return err
		}
		prefix, err := s.DAO.ByNameTx(tx, accessionNumberPrefix)
		if err != nil || prefix == nil || strings.TrimSpace(derefStr(prefix.Value)) != "" {
			return err
		}
		// The ACTING user, not a constant. Java carries the id forward with
		// setSysUserId(siteInformation.getSysUserId()), so the prefix row's
		// audit entry names whoever changed the site number. The hardcoded 1
		// this replaced happened to be right for admin and wrong for everyone
		// else.
		return s.DAO.UpdateValueTx(tx, strconv.FormatInt(prefix.ID, 10), &value, sysUserID)
	}
	return nil
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
func (s *SiteInformationService) Delete(ids []string, sysUserID int64) error {
	if err := s.DAO.DeleteAll(ids, sysUserID); err != nil {
		return err
	}
	return s.Reload()
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

// Locale is LocaleContextHolder.getLocale(), read from the CONFIGURATION
// CACHE rather than from a value captured at startup.
//
// Writing the "default language locale" row changes it, and the change is
// visible on the very next request: the write reloads the cache, and
// localeResolver.setLocale puts the same locale on the session. Measured — with
// a French text seeded into a localization, flipping the row to fr-FR turned
// localizedValue from "Test LIMS" into the French value and the display-language
// names from "English:"/"French:" into "anglais:"/"français:".
//
// The stored value is a language TAG ("en-US"); Locale.forLanguageTag(...)
// .getLanguage() reduces it to the language, which is what localization_value
// is keyed by.
func (s *SiteInformationService) Locale() string {
	value, ok := s.configValue("default language locale")
	if !ok {
		if s.ActiveLocale != "" {
			return s.ActiveLocale
		}
		return "en"
	}
	if i := strings.IndexAny(value, "-_"); i > 0 {
		return value[:i]
	}
	return value
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// LabUnitConfig ports getLabUnitConfig — four ConfigurationProperties values
// assembled into a HashMap.
//
// labName is ABSENT from the response, not empty: Property.SiteName has no
// site_information row in this deployment, getPropertyValue returns null, and
// Include.NON_NULL drops the key. A port that emitted "" would add a field Java
// does not send.
func (s *SiteInformationService) LabUnitConfig() (map[string]any, error) {
	config := map[string]any{}

	// The only one with a fallback: an unset workflow type reads as "Both".
	workflow, ok := s.configValue("orderEntryWorkflowType")
	if !ok {
		workflow = "Both"
	}
	config["workflowType"] = workflow

	// labName is dropped when the value is BLANK, not only when the row is
	// missing — configValue treats the two the same because Java does.
	if labName, ok := s.configValue("SiteName"); ok {
		config["labName"] = labName
	}

	// isPropertyValueEqual, so unset is false rather than absent.
	validate, _ := s.configValue("validateAccessionNumber")
	config["useAccessionNumberValidation"] = validate == "true"

	if format, ok := s.configValue("acessionFormat"); ok {
		config["accessionFormat"] = format
	}
	return config, nil
}

// Reload rebuilds the configuration snapshot.
//
// ConfigurationProperties is a CACHE, not a view: Java loads every
// site_information row into memory at startup and reloads it only after a write
// of its own. A row changed in the database by anything else is invisible until
// then — measured, by editing acessionFormat directly and watching Java keep
// answering the old value while a live-reading port answered the new one.
//
// So the port caches too. Reading the table on every request would be more
// correct and less faithful, and the whole point of this exercise is the second
// one.
func (s *SiteInformationService) Reload() error {
	rows, err := s.DAO.All()
	if err != nil {
		return err
	}
	s.configMu.Lock()
	s.config = rows
	s.configMu.Unlock()
	return nil
}

// configValue is ConfigurationProperties.getPropertyValue: the cached value,
// and whether the property is set at all. A BLANK value reads as unset —
// site_information row 33 is SiteName with an empty column, and Java omits
// labName rather than sending "".
func (s *SiteInformationService) configValue(name string) (string, bool) {
	s.configMu.RLock()
	defer s.configMu.RUnlock()
	v, ok := s.config[name]
	if !ok || v == "" {
		return "", false
	}
	return v, true
}
