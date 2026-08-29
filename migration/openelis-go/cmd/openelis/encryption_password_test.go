package main

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"openelis-go/internal/security/encryption"
)

// The Go service and the Java webapp share one database, and a site_information
// row flagged `encrypted` holds jasypt ciphertext keyed on
// encryption.general.password. If the two processes resolve that property to
// different strings, neither can read what the other wrote — and nothing fails
// at startup to say so. The first symptom is a 500 on the config form for a row
// that looks fine in the table.
//
// So the compose file that runs them side by side has to carry the key, and it
// has to be the SAME key. That is what these two tests check; there is nothing
// to assert at runtime, because the failure is a silent divergence rather than
// an error.

const (
	composePath    = "../../docker-compose.go.yml"
	propertiesPath = "../../../../volume/properties/common.properties"
)

// javaEncryptionPassword is what the Java webapp resolves
// encryption.general.password to in this deployment.
func javaEncryptionPassword(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(propertiesPath)
	if err != nil {
		t.Skipf("deployment properties not readable from here: %v", err)
	}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "encryption.general.password=") {
			return strings.TrimSpace(strings.TrimPrefix(line, "encryption.general.password="))
		}
	}
	t.Fatalf("%s does not set encryption.general.password", propertiesPath)
	return ""
}

// The variable must be there at all. Without it the service silently takes
// Spring's `dev` default, which is not this deployment's key.
func TestComposePassesEncryptionPassword(t *testing.T) {
	raw, err := os.ReadFile(composePath)
	if err != nil {
		t.Skipf("compose file not readable from here: %v", err)
	}
	if !strings.Contains(string(raw), "OE_ENCRYPTION_PASSWORD") {
		t.Fatalf("%s does not pass OE_ENCRYPTION_PASSWORD; the service would fall back to %q "+
			"while the Java container beside it uses the deployment key",
			composePath, encryption.DefaultPassword)
	}
}

// And it must match Java. A key that is present but different fails exactly the
// same way as one that is absent.
func TestComposeEncryptionPasswordMatchesJava(t *testing.T) {
	raw, err := os.ReadFile(composePath)
	if err != nil {
		t.Skipf("compose file not readable from here: %v", err)
	}
	// `- OE_ENCRYPTION_PASSWORD=${OE_ENCRYPTION_PASSWORD:-kspass}` — the default
	// after `:-` is what an operator who sets nothing gets, so that is the value
	// under test.
	re := regexp.MustCompile(`OE_ENCRYPTION_PASSWORD=(?:\$\{OE_ENCRYPTION_PASSWORD:-([^}]*)\}|([^\s#]+))`)
	m := re.FindStringSubmatch(string(raw))
	if m == nil {
		t.Fatalf("%s sets no value for OE_ENCRYPTION_PASSWORD", composePath)
	}
	got := m[1]
	if got == "" {
		got = m[2]
	}

	want := javaEncryptionPassword(t)
	if got != want {
		t.Errorf("compose gives Go %q; the Java webapp beside it uses %q. "+
			"Rows encrypted by one are unreadable by the other.", got, want)
	}
}

// The fallback is still the right default for a Go service running ALONE
// against a database Java never touched — this is a deployment-wiring bug, not
// a reason to hardcode a deployment's key into the binary.
func TestEncryptionPasswordFallsBackToSpringDefault(t *testing.T) {
	t.Setenv("OE_ENCRYPTION_PASSWORD", "")
	if got := encryptionPassword(); got != encryption.DefaultPassword {
		t.Errorf("unset OE_ENCRYPTION_PASSWORD gave %q, want the Spring default %q",
			got, encryption.DefaultPassword)
	}
	t.Setenv("OE_ENCRYPTION_PASSWORD", "from-the-environment")
	if got := encryptionPassword(); got != "from-the-environment" {
		t.Errorf("set OE_ENCRYPTION_PASSWORD gave %q, want the value from the environment", got)
	}
}
