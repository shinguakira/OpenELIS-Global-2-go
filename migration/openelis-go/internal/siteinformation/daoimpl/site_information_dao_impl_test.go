package daoimpl

import "testing"

// The delete payload is a wire contract: the audit trail is read back by
// SystemAuditEventRestController, so a field this port drops is a field the
// history screen can never show again.
//
// Both cases below are measured off the live Java server through
// tests/mutating/e1-config-parity-gaps.spec.ts. These tests exist so a change
// here fails without a Postgres, a Tomcat and a login.

func str(s string) *string { return &s }

// The shape every row the insert path creates has: no tag, no keys, no
// dictionary category. This is what the first port hardcoded, and for these
// rows it was right.
func TestDeleteChanges_bareRow(t *testing.T) {
	got := deleteChanges(doomed{
		ID:          "999",
		Name:        "e2eConfigProbe",
		Description: "probe",
		Value:       "beta",
		Encrypted:   false,
		ValueType:   "text",
		Domain:      "siteIdentity",
		Group:       0,
		Schedule:    "",
		Lastupdated: str("2026-08-29 07:20:16.474"),
	})

	want := "<name>e2eConfigProbe</name>\n" +
		"<description>probe</description>\n" +
		"<value>beta</value>\n" +
		"<encrypted>false</encrypted>\n" +
		"<valueType>text</valueType>\n" +
		"<domain>siteIdentity</domain>\n" +
		"<group>0</group>\n" +
		"<schedule></schedule>\n" +
		"<lastupdated>2026-08-29 07:20:16.474</lastupdated>\n"
	if got != want {
		t.Errorf("payload for a bare row\n got: %q\nwant: %q", got, want)
	}
}

// A row that carries the optional columns — the shipped bannerHeading row (82)
// is one, and it is selectable for deletion like any other.
//
// The ORDER is the point, not just the presence: getChanges walks the entity's
// declared fields, so instructionKey lands between valueType and domain, tag
// between group and schedule, and the two after schedule follow the same rule.
// Appending them at the end would carry every value and still not match.
func TestDeleteChanges_populatedOptionalColumns(t *testing.T) {
	got := deleteChanges(doomed{
		ID:                   "82",
		Name:                 "bannerHeading",
		Description:          "Text for banner",
		Value:                "2",
		Encrypted:            false,
		ValueType:            "text",
		InstructionKey:       "instructions.bannerHeading",
		Domain:               "siteIdentity",
		Group:                0,
		Tag:                  "localization",
		Schedule:             "",
		DictionaryCategoryID: "197",
		DescriptionKey:       "siteInfo.bannerHeading",
		Lastupdated:          str("2020-01-23 05:46:11.082"),
	})

	want := "<name>bannerHeading</name>\n" +
		"<description>Text for banner</description>\n" +
		"<value>2</value>\n" +
		"<encrypted>false</encrypted>\n" +
		"<valueType>text</valueType>\n" +
		"<instructionKey>instructions.bannerHeading</instructionKey>\n" +
		"<domain>siteIdentity</domain>\n" +
		"<group>0</group>\n" +
		"<tag>localization</tag>\n" +
		"<schedule></schedule>\n" +
		"<dictionaryCategoryId>197</dictionaryCategoryId>\n" +
		"<descriptionKey>siteInfo.bannerHeading</descriptionKey>\n" +
		"<lastupdated>2020-01-23 05:46:11.082</lastupdated>\n"
	if got != want {
		t.Errorf("payload for a row with optional columns\n got: %q\nwant: %q", got, want)
	}
}

// One populated column does not drag the others in: getChanges emits a field
// only when it differs from the blank object, so each is independent.
func TestDeleteChanges_onlyPopulatedOnesAppear(t *testing.T) {
	got := deleteChanges(doomed{
		Name:      "half",
		ValueType: "text",
		Domain:    "siteIdentity",
		Tag:       "enable",
	})
	for _, absent := range []string{"instructionKey", "dictionaryCategoryId", "descriptionKey"} {
		if contains(got, "<"+absent+">") {
			t.Errorf("empty %s was emitted: %q", absent, got)
		}
	}
	if !contains(got, "<tag>enable</tag>") {
		t.Errorf("populated tag was dropped: %q", got)
	}
}

// escapeXML runs on these too — a description with an ampersand is not
// hypothetical, and an unescaped one makes the payload unparseable.
func TestDeleteChanges_escapesOptionalColumns(t *testing.T) {
	got := deleteChanges(doomed{Name: "x", Tag: "a&b<c>"})
	if !contains(got, "<tag>a&amp;b&lt;c&gt;</tag>") {
		t.Errorf("tag was not escaped: %q", got)
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
