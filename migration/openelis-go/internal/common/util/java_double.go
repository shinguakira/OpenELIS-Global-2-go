package util

import (
	"math"
	"strconv"
	"strings"
)

// JavaDouble marshals a float64 the way Jackson serialises a java.lang.Double,
// which is NOT how Go's encoding/json serialises float64.
//
// Go writes the shortest round-tripping form and drops a trailing ".0", so a
// quantity of 5.0 goes out as `5`. Java's Double.toString always keeps at least
// one digit after the point, so the same column goes out as `5.0`. The values
// parse equal, which is exactly why this survived every JSON-level comparison —
// a differ that unmarshals both sides sees 5 == 5.0 and reports parity. It is
// visible only in the raw bytes, and it changes Content-Length.
//
// Double.toString's shape:
//   - 1e-3 <= |d| < 1e7  -> plain decimal, at least one fraction digit
//   - otherwise          -> computerized scientific notation, "1.0E20",
//     mantissa in [1,10) with at least one fraction digit, and an exponent
//     with no '+' and no leading zeros
//
// The scientific branch is unreachable for the columns this port serialises
// (quantities and GPS coordinates), but it is the documented contract and
// cheaper to write than to justify omitting.
type JavaDouble float64

// MarshalJSON implements json.Marshaler.
func (d JavaDouble) MarshalJSON() ([]byte, error) {
	return []byte(JavaDoubleString(float64(d))), nil
}

// JavaDoubleString renders v as Java's Double.toString would.
func JavaDoubleString(v float64) string {
	// NaN and the infinities have no JSON representation and Jackson refuses
	// them outright; no column this port reads can hold one. Rendering them as
	// null keeps the output valid JSON instead of emitting a bare NaN token.
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return "null"
	}
	if v == 0 {
		if math.Signbit(v) {
			return "-0.0"
		}
		return "0.0"
	}

	if a := math.Abs(v); a >= 1e-3 && a < 1e7 {
		s := strconv.FormatFloat(v, 'f', -1, 64)
		if !strings.Contains(s, ".") {
			s += ".0"
		}
		return s
	}

	// strconv gives "1E+20" / "1.234E-05"; Java gives "1.0E20" / "1.234E-5".
	s := strconv.FormatFloat(v, 'E', -1, 64)
	mantissa, exponent, found := strings.Cut(s, "E")
	if !found {
		return s
	}
	if !strings.Contains(mantissa, ".") {
		mantissa += ".0"
	}
	negative := strings.HasPrefix(exponent, "-")
	exponent = strings.TrimLeft(strings.TrimLeft(exponent, "+-"), "0")
	if exponent == "" {
		exponent = "0"
	}
	if negative {
		exponent = "-" + exponent
	}
	return mantissa + "E" + exponent
}

// JavaDoublePtr adapts a nullable DAO column to the wire type. Returns nil for
// nil so Include.NON_NULL semantics are preserved.
func JavaDoublePtr(v *float64) *JavaDouble {
	if v == nil {
		return nil
	}
	d := JavaDouble(*v)
	return &d
}
