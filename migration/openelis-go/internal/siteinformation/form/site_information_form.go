// Package form ports org.openelisglobal.siteinformation.form.
// Folder layout mirrors the Java source during migration.
//
// Two shapes, and they are NOT the same object even though one endpoint pair
// produces both — see SiteInformationForm's doc comment.
package form

import locform "openelis-go/internal/localization/form"

// SiteInformationForm is what GET /rest/{domain} returns.
//
// Field ORDER is the wire contract: Jackson serialises a bean in declaration
// order, BaseForm's fields first, so formName…cancelMethod precede the
// SiteInformationForm fields.
type SiteInformationForm struct {
	FormName string `json:"formName"`
	// FormAction is set only by setupFormForRequest, which the GET path calls
	// and the POST path does not — hence omitempty, and hence the key's absence
	// from every POST response.
	FormAction     string `json:"formAction,omitempty"`
	FormMethod     string `json:"formMethod"`
	CancelAction   string `json:"cancelAction"`
	SubmitOnCancel bool   `json:"submitOnCancel"`
	CancelMethod   string `json:"cancelMethod"`

	ParamName   string `json:"paramName"`
	Description string `json:"description"`
	Value       string `json:"value"`
	Encrypted   bool   `json:"encrypted"`
	ValueType   string `json:"valueType"`
	// SiteInfoDomainName is the DOMAIN LABEL, and it is not the path segment:
	// /PatientConfiguration answers "PaitientConfiguration", /SampleEntryConfig
	// answers "sampleEntryConfig", /NonConformityConfiguration answers
	// "non_conformityConfiguration". The menu controller spells the first of
	// those correctly, so one domain has two names depending on which endpoint
	// is asked. All of it is the wire contract; none of it is tidied here.
	SiteInfoDomainName string `json:"siteInfoDomainName"`

	// DictionaryValues is present only for a dictionary-typed row.
	DictionaryValues []string `json:"dictionaryValues,omitempty"`

	Editable bool `json:"editable"`
	// Tag is a POINTER because the key's presence is observable: the blank form
	// carries "" and a loaded row whose tag column is NULL drops the key under
	// Include.NON_NULL. Same endpoint, same field, present or absent depending
	// on which branch ran.
	Tag            *string `json:"tag,omitempty"`
	DescriptionKey string  `json:"descriptionKey"`
}

// NewSiteInformationForm reproduces `new SiteInformationForm()` — the bean
// initialisers plus the constructor's setFormName.
//
// These defaults are not decoration: the POST handler never calls
// setupFormForRequest, so its response is the deserialised request body over
// exactly this object, which is why a POST answers formName
// "siteInformationForm" where the GET answers "SiteInformationForm", and
// cancelAction "Home" where the GET answers "CancelSiteInformation".
func NewSiteInformationForm() SiteInformationForm {
	empty := ""
	return SiteInformationForm{
		FormName:       "siteInformationForm",
		FormMethod:     "POST",
		CancelAction:   "Home",
		SubmitOnCancel: false,
		CancelMethod:   "POST",
		ParamName:      "",
		Description:    "",
		Value:          "",
		ValueType:      "text",
		Editable:       true,
		Tag:            &empty,
		DescriptionKey: "",
	}
}

// SiteInformationPost is the request body of POST /rest/{domain}.
//
// The row id is NOT in here — it arrives as the `ID` query parameter. A body
// that carries an id is ignored.
type SiteInformationPost struct {
	ParamName          *string `json:"paramName"`
	Description        *string `json:"description"`
	Value              *string `json:"value"`
	Encrypted          *bool   `json:"encrypted"`
	ValueType          *string `json:"valueType"`
	SiteInfoDomainName *string `json:"siteInfoDomainName"`
	Editable           *bool   `json:"editable"`
	Tag                *string `json:"tag"`
	DescriptionKey     *string `json:"descriptionKey"`
}

// SiteInformationMenuForm is what GET /rest/{domain}Menu returns.
type SiteInformationMenuForm struct {
	FormName       string `json:"formName"`
	FormMethod     string `json:"formMethod"`
	CancelAction   string `json:"cancelAction"`
	SubmitOnCancel bool   `json:"submitOnCancel"`
	CancelMethod   string `json:"cancelMethod"`

	// The paging block is DECORATIVE. Measured across five domains holding 0,
	// 3, 4, 13 and 30 rows: every one answers total "", from "1", to "20", and
	// menuList carries the whole list regardless. createMenuList calls
	// setDisplayPageBounds with the full size and the client paginates.
	TotalRecordCount string `json:"totalRecordCount"`
	FromRecordCount  string `json:"fromRecordCount"`
	ToRecordCount    string `json:"toRecordCount"`

	MenuList           []MenuItem `json:"menuList"`
	SelectedIDs        []string   `json:"selectedIDs"`
	SiteInfoDomainName string     `json:"siteInfoDomainName"`
}

// MenuItem is one serialised SiteInformation inside menuList.
type MenuItem struct {
	Lastupdated    *int64     `json:"lastupdated,omitempty"`
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Description    *string    `json:"description,omitempty"`
	Value          *string    `json:"value,omitempty"`
	Encrypted      bool       `json:"encrypted"`
	ValueType      string     `json:"valueType"`
	InstructionKey *string    `json:"instructionKey,omitempty"`
	Domain         *DomainDTO `json:"domain,omitempty"`
	Group          int        `json:"group"`
	Tag            *string    `json:"tag,omitempty"`
	// Localization is the FULL serialised Localization graph, nested inside the
	// menu row. It appears only for a row tagged "localization", where the
	// site_information value column holds a localization id rather than a value.
	Localization *locform.LocalizationDTO `json:"localization,omitempty"`
	// DictionaryCategoryID sits between localization and descriptionKey because
	// that is where the Java bean declares it — with the unserialised schedule
	// holder in between.
	DictionaryCategoryID *string `json:"dictionaryCategoryId,omitempty"`
	DescriptionKey       *string `json:"descriptionKey,omitempty"`
}

// DomainDTO is the nested site_information_domain object.
type DomainDTO struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

// DeleteRequest is the body of GET /rest/Delete{domain} — a GET that reads a
// request body, which is unusual enough to be worth naming.
type DeleteRequest struct {
	SelectedIDs        []string `json:"selectedIDs"`
	SiteInfoDomainName *string  `json:"siteInfoDomainName"`
}

// ObjectError is one entry of the bare ARRAY Spring returns when
// Delete{domain}'s form fails validation — a fourth error envelope, alongside
// the RFC 7807 ProblemDetail, the per-field `errors` map, and Tomcat's
// {timestamp,status,error}.
type ObjectError struct {
	Codes          []string `json:"codes"`
	Arguments      []string `json:"arguments"`
	DefaultMessage string   `json:"defaultMessage"`
	ObjectName     string   `json:"objectName"`
	Field          string   `json:"field"`
	// RejectedValue is the value that failed — a FieldError carries it where a
	// plain ObjectError does not, and it lands between field and bindingFailure.
	RejectedValue  string `json:"rejectedValue"`
	BindingFailure bool   `json:"bindingFailure"`
	Code           string `json:"code"`
}

// ProblemDetail is Spring's RFC 7807 body. As everywhere else in this port the
// four text fields are the UNRESOLVED message keys, because no MessageSource is
// wired for ProblemDetail — see java-possible-bugs.md C5.
type ProblemDetail struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail"`
	Instance string `json:"instance"`
}

const (
	notReadableException  = "org.springframework.http.converter.HttpMessageNotReadableException"
	mediaTypeNotSupported = "org.springframework.web.HttpMediaTypeNotSupportedException"
)

// NotReadableProblem is the 400 for a body Jackson cannot parse — malformed
// JSON on the POST, or no body at all on the delete, which has no `consumes`
// clause to reject it earlier.
func NotReadableProblem(path string) ProblemDetail {
	return problem(notReadableException, 400, path)
}

// UnsupportedMediaProblem is the 415 the POST answers when the request carries
// no JSON content type. `consumes = APPLICATION_JSON_VALUE` on that mapping
// rejects it before the body is ever read, which is why the same empty request
// is a 415 here and a 400 on the delete.
func UnsupportedMediaProblem(path string) ProblemDetail {
	return problem(mediaTypeNotSupported, 415, path)
}

func problem(exception string, status int, path string) ProblemDetail {
	return ProblemDetail{
		Type:   "problemDetail.type." + exception,
		Title:  "problemDetail.title." + exception,
		Status: status,
		Detail: "problemDetail." + exception,
		// The request path as it arrived, context prefix included — the Go
		// service answers on both the bare path and /api/OpenELIS-Global.
		Instance: path,
	}
}
