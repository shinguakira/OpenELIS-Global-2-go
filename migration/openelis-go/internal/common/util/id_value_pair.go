// Package util ports org.openelisglobal.common.util — shared value types.
// Folder layout mirrors the Java source during migration; idiomatic Go reorg
// comes at the end.
package util

// IdValuePair mirrors org.openelisglobal.common.util.IdValuePair. It serializes
// as {"id":...,"value":...}; field order is fixed to match Jackson's output.
type IdValuePair struct {
	Id    string `json:"id"`
	Value string `json:"value"`
}

// NewIdValuePair is the analog of the Java constructor new IdValuePair(id, value).
func NewIdValuePair(id, value string) IdValuePair {
	return IdValuePair{Id: id, Value: value}
}
