package googlecsv

import (
	"strings"
	"testing"
)

func TestDetectPasskeyHeaderRecognizesKnownColumns(t *testing.T) {
	cases := [][]string{
		{"name", "url", "credential_id"},
		{"name", "url", "CredentialID"},
		{"name", "passkey"},
		{"name", "public_key_credential", "username"},
	}
	for _, header := range cases {
		if !DetectPasskeyHeader(header) {
			t.Errorf("DetectPasskeyHeader(%v) = false, want true", header)
		}
	}
}

func TestDetectPasskeyHeaderFalseForPasswordHeader(t *testing.T) {
	header := []string{"name", "url", "username", "password", "note"}
	if DetectPasskeyHeader(header) {
		t.Errorf("DetectPasskeyHeader(%v) = true, want false", header)
	}
}

func TestParseSetsPasskeyCountForPasskeyShapedFile(t *testing.T) {
	csvData := "name,url,username,credential_id\n" +
		"GitHub,https://github.com,octocat,abc123\n" +
		"GitLab,https://gitlab.com,octocat,def456\n" +
		"Bitbucket,https://bitbucket.org,octocat,ghi789\n"

	result, err := Parse(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if !result.UnknownSchema {
		t.Error("UnknownSchema = false, want true for a passkey-shaped header")
	}
	if result.PasskeyCount != 3 {
		t.Errorf("PasskeyCount = %d, want 3", result.PasskeyCount)
	}
	if len(result.Entries) != 0 {
		t.Errorf("Entries = %d, want 0 (passkeys are not imported)", len(result.Entries))
	}
}

func TestParseLeavesPasskeyCountZeroForNormalPasswordFile(t *testing.T) {
	csvData := "name,url,username,password\nGitHub,https://github.com,octocat,hunter2\n"

	result, err := Parse(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if result.PasskeyCount != 0 {
		t.Errorf("PasskeyCount = %d, want 0", result.PasskeyCount)
	}
	if result.UnknownSchema {
		t.Error("UnknownSchema = true, want false for a normal password header")
	}
}

func TestPasskeyWarningMatchesSpecExample(t *testing.T) {
	want := "Unsupported credentials detected: 3 passkeys.\nThey were not imported and remain outside this vault."
	if got := PasskeyWarning(3); got != want {
		t.Errorf("PasskeyWarning(3) = %q, want %q", got, want)
	}
}

func TestPasskeyWarningSingularHasNoTrailingS(t *testing.T) {
	if got := PasskeyWarning(1); !strings.Contains(got, "1 passkey.") {
		t.Errorf("PasskeyWarning(1) = %q, want it to contain %q", got, "1 passkey.")
	}
}
