// Package i18n bundles the Java message_en.properties file and exposes
// a parsed key→value map. Mirrors Spring's ResourceBundleMessageSource
// for the English baseline used by all ported endpoints that return
// localized labels (status types, gender, sample types, etc.).
//
// The .properties file is a copy of
//
//	src/main/resources/languages/message_en.properties
//
// kept in sync with the Java source tree. If a key drifts the parity e2e
// test will catch it (exact-equality assertions on the captured baseline).
package i18n

import (
	"bufio"
	_ "embed"
	"strings"
)

//go:embed message_en.properties
var messageEnSrc string

// Messages parses and returns the bundled message_en.properties as a
// key→value map. Call once at startup and pass the result to services
// that need label resolution.
func Messages() map[string]string {
	return parseProperties(messageEnSrc)
}

func parseProperties(src string) map[string]string {
	m := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(src))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line[0] == '#' || line[0] == '!' {
			continue
		}

		// LINE CONTINUATIONS. java.util.Properties treats a trailing backslash
		// as "this value continues on the next line", joining them and dropping
		// the continuation's leading whitespace. Without this,
		// label.select.patient.ID reads "Patient identification \" instead of
		// "Patient identification code" — the value is silently truncated at the
		// backslash and the label reaches the UI mangled.
		for strings.HasSuffix(line, `\`) && scanner.Scan() {
			line = strings.TrimSuffix(line, `\`) + strings.TrimSpace(scanner.Text())
		}

		idx := strings.IndexByte(line, '=')
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		m[key] = val
	}
	return m
}
