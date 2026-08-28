// Package form ports the response shapes of
// org.openelisglobal.genericsample.controller.rest.GenericSampleOrderRestController.
// Folder layout mirrors the Java source during migration.
package form

import "net/http"

// ErrorDTO mirrors the handler's own Map.of("error", "...") envelope — ONE key.
// Distinct from the ProblemDetail Spring produces for a binding failure on the
// same endpoint.
type ErrorDTO struct {
	Error string `json:"error"`
}

// ProblemDetail mirrors Spring's RFC 7807 envelope as this deployment emits it.
//
// The four message fields come back as UNRESOLVED message keys
// ("problemDetail.title.org.springframework...") rather than English sentences,
// because no MessageSource is wired for ProblemDetail on this app. Reproduced
// as-is: a port that emitted readable text would be a different response, and
// the frontend keys off the type.
type ProblemDetail struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail"`
	Instance string `json:"instance"`
}

// missingParamException is the exception class Spring names in all four fields.
const missingParamException = "org.springframework.web.bind.MissingServletRequestParameterException"

// MissingAccessionNumberProblem builds the 400 body for a missing
// accessionNumber.
//
// `instance` is the REQUEST PATH — including the servlet context prefix when the
// request arrived on it, which is why it is derived from the request rather than
// hardcoded. Java is deployed under /api/OpenELIS-Global and emits
// "/api/OpenELIS-Global/rest/GenericSampleOrder"; the Go service answers on both
// that prefix and the bare path, so it must echo whichever one was used.
func MissingAccessionNumberProblem(r *http.Request) ProblemDetail {
	return ProblemDetail{
		Type:     "problemDetail.type." + missingParamException,
		Title:    "problemDetail.title." + missingParamException,
		Status:   http.StatusBadRequest,
		Detail:   "problemDetail." + missingParamException,
		Instance: r.URL.Path,
	}
}
