// Package valueholder ports org.openelisglobal.testcalculated.valueholder.
// Folder layout mirrors the Java source during migration.
package valueholder

import "openelis-go/internal/common/util"

// MathFunctions mirrors Operation.mathFunctions() (Operation.java:135). This is
// the fixed operator catalog returned by GET /rest/math-functions — a compiled-in
// constant in Java, reproduced here byte-for-byte (14 entries, same order).
func MathFunctions() []util.IdValuePair {
	return []util.IdValuePair{
		util.NewIdValuePair("+", "Plus"),
		util.NewIdValuePair("-", "Minus"),
		util.NewIdValuePair("/", "Divided By"),
		util.NewIdValuePair("*", "Multiplied By"),
		util.NewIdValuePair("(", "Open Bracket"),
		util.NewIdValuePair(")", "Close Bracket"),
		util.NewIdValuePair("==", "Equals"),
		util.NewIdValuePair("!=", "Does Not Equal"),
		util.NewIdValuePair(">=", "Is Greater Than Or Equal"),
		util.NewIdValuePair("<=", "Is Less Than Or Equal"),
		util.NewIdValuePair("IS_IN_NORMAL_RANGE", "Is With In Normal Range"),
		util.NewIdValuePair("IS_OUTSIDE_NORMAL_RANGE", "Is Out Side Normal Range"),
		util.NewIdValuePair("&&", "And"),
		util.NewIdValuePair("||", "Or"),
	}
}
