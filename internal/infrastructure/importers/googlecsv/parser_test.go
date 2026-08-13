package googlecsv

import (
	"strings"
	"testing"
)

func revealPassword(t *testing.T, e ImportedEntry) string {
	t.Helper()
	var got []byte
	if err := e.Password.Reveal(func(value []byte) error {
		got = append(got, value...)
		return nil
	}); err != nil {
		t.Fatalf("Reveal() error = %v", err)
	}
	return string(got)
}

func TestParseBasicCSV(t *testing.T) {
	csv := "name,url,username,password,note\n" +
		"GitHub,https://github.com,octocat,s3cr3t,dev account\n" +
		"AWS,https://aws.amazon.com,alice,another-pass,\n"

	result, err := Parse(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if result.UnknownSchema {
		t.Fatal("UnknownSchema = true, want false")
	}
	if len(result.Entries) != 2 {
		t.Fatalf("len(Entries) = %d, want 2", len(result.Entries))
	}

	first := result.Entries[0]
	if first.Title != "GitHub" || first.Username != "octocat" || first.URL != "https://github.com" || first.Notes != "dev account" {
		t.Errorf("first entry = %+v", first)
	}
	if got := revealPassword(t, first); got != "s3cr3t" {
		t.Errorf("first password = %q, want %q", got, "s3cr3t")
	}
	if first.SourceRow != 2 {
		t.Errorf("first SourceRow = %d, want 2 (header is row 1)", first.SourceRow)
	}
	if result.Entries[1].SourceRow != 3 {
		t.Errorf("second SourceRow = %d, want 3", result.Entries[1].SourceRow)
	}
}

func TestParseStripsUTF8BOM(t *testing.T) {
	// Built from an explicit byte escape rather than pasting a literal
	// BOM character into this source file: Go's lexer only tolerates a
	// BOM at the very start of a file and rejects one appearing mid-file,
	// which a literal BOM character here would be.
	csv := "\xEF\xBB\xBFname,url,username,password,note\nGitHub,https://github.com,octocat,pass,\n"

	result, err := Parse(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if result.UnknownSchema {
		t.Fatal("UnknownSchema = true, want false (BOM should not corrupt the first header name)")
	}
	if len(result.Entries) != 1 {
		t.Fatalf("len(Entries) = %d, want 1", len(result.Entries))
	}
}

func TestParseDetectsSemicolonDelimiter(t *testing.T) {
	csv := "name;url;username;password;note\nGitHub;https://github.com;octocat;pass;\n"

	result, err := Parse(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(result.Entries) != 1 {
		t.Fatalf("len(Entries) = %d, want 1", len(result.Entries))
	}
	if result.Entries[0].Title != "GitHub" {
		t.Errorf("Title = %q, want %q", result.Entries[0].Title, "GitHub")
	}
}

func TestParseDetectsTabDelimiter(t *testing.T) {
	csv := "name\turl\tusername\tpassword\tnote\nGitHub\thttps://github.com\toctocat\tpass\t\n"

	result, err := Parse(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(result.Entries) != 1 {
		t.Fatalf("len(Entries) = %d, want 1", len(result.Entries))
	}
}

func TestParseHandlesQuotedCommaAndMultilineField(t *testing.T) {
	csv := "name,url,username,password,note\n" +
		"\"Client, Inc.\",https://example.com,alice,pass,\"line one\nline two\"\n"

	result, err := Parse(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(result.Entries) != 1 {
		t.Fatalf("len(Entries) = %d, want 1", len(result.Entries))
	}
	e := result.Entries[0]
	if e.Title != "Client, Inc." {
		t.Errorf("Title = %q, want %q", e.Title, "Client, Inc.")
	}
	if e.Notes != "line one\nline two" {
		t.Errorf("Notes = %q, want %q", e.Notes, "line one\nline two")
	}
}

func TestParseMapsKnownHeaderAliases(t *testing.T) {
	csv := "title,login_uri,login_username,login_password,notes\n" +
		"GitHub,https://github.com,octocat,pass,\n"

	result, err := Parse(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if result.UnknownSchema {
		t.Fatal("UnknownSchema = true, want false")
	}
	if len(result.Entries) != 1 {
		t.Fatalf("len(Entries) = %d, want 1", len(result.Entries))
	}
	if result.Entries[0].URL != "https://github.com" || result.Entries[0].Username != "octocat" {
		t.Errorf("entry = %+v", result.Entries[0])
	}
}

func TestParseReturnsUnknownSchemaWhenHeadersUnrecognized(t *testing.T) {
	csv := "foo,bar,baz\n1,2,3\n"

	result, err := Parse(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if !result.UnknownSchema {
		t.Error("UnknownSchema = false, want true")
	}
	if len(result.Entries) != 0 {
		t.Errorf("len(Entries) = %d, want 0", len(result.Entries))
	}
}

func TestParseSkipsRowsExceedingMaxFieldSize(t *testing.T) {
	oversized := strings.Repeat("a", MaxFieldBytes+1)
	csv := "name,url,username,password,note\n" +
		"Good,https://example.com,alice,pass,\n" +
		"Bad," + oversized + ",alice,pass,\n" +
		"AlsoGood,https://example.org,bob,pass2,\n"

	result, err := Parse(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if result.InvalidRows != 1 {
		t.Errorf("InvalidRows = %d, want 1", result.InvalidRows)
	}
	if len(result.Entries) != 2 {
		t.Fatalf("len(Entries) = %d, want 2 (the oversized row should be skipped, not abort parsing)", len(result.Entries))
	}
	if result.Entries[0].Title != "Good" || result.Entries[1].Title != "AlsoGood" {
		t.Errorf("entries = %+v", result.Entries)
	}
}

func TestParseCountsRowsMissingTitleAsInvalid(t *testing.T) {
	csv := "name,url,username,password,note\n" +
		",https://example.com,alice,pass,\n" + // no title
		"Good,https://example.org,bob,pass2,\n"

	result, err := Parse(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if result.InvalidRows != 1 {
		t.Errorf("InvalidRows = %d, want 1", result.InvalidRows)
	}
	if len(result.Entries) != 1 || result.Entries[0].Title != "Good" {
		t.Errorf("Entries = %+v", result.Entries)
	}
}

func TestParseSanitizesPreviewableFields(t *testing.T) {
	csv := "name,url,username,password,note\n" +
		"Evil\x1b[31mTitle,https://example.com,user\x07name,pass,note\x1btext\n"

	result, err := Parse(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(result.Entries) != 1 {
		t.Fatalf("len(Entries) = %d, want 1", len(result.Entries))
	}
	e := result.Entries[0]
	for _, field := range []string{e.Title, e.Username, e.Notes} {
		if strings.ContainsAny(field, "\x1b\x07") {
			t.Errorf("field %q still contains an unsanitized control character", field)
		}
	}
}

func TestParsePasswordIsNotSanitized(t *testing.T) {
	// The password must survive import as the exact credential bytes.
	// This value contains a tab, which terminal.Sanitize would normally
	// pass through unchanged anyway, but the point is that Password never
	// even goes through Sanitize in the first place — verified against
	// the exact byte sequence.
	rawPassword := "p@ss\tword!"
	csv := "name,url,username,password,note\n" +
		"GitHub,https://github.com,octocat," + rawPassword + ",\n"

	result, err := Parse(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(result.Entries) != 1 {
		t.Fatalf("len(Entries) = %d, want 1", len(result.Entries))
	}
	if got := revealPassword(t, result.Entries[0]); got != rawPassword {
		t.Errorf("password = %q, want unmodified %q", got, rawPassword)
	}
}

func TestParseDoesNotInterpretFormulas(t *testing.T) {
	csv := "name,url,username,password,note\n" +
		"=cmd|' /C calc'!A1,https://example.com,alice,pass,=1+1\n"

	result, err := Parse(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(result.Entries) != 1 {
		t.Fatalf("len(Entries) = %d, want 1", len(result.Entries))
	}
	e := result.Entries[0]
	if e.Title != "=cmd|' /C calc'!A1" {
		t.Errorf("Title = %q, want the literal formula-looking string unevaluated", e.Title)
	}
	if e.Notes != "=1+1" {
		t.Errorf("Notes = %q, want the literal string %q", e.Notes, "=1+1")
	}
}

func TestParseEmptyFileReturnsNoHeaderError(t *testing.T) {
	_, err := Parse(strings.NewReader(""))
	if err == nil {
		t.Fatal("Parse() error = nil, want ErrNoHeader")
	}
}

func TestParseMultipleInvalidRowsDoNotAbortParsing(t *testing.T) {
	oversized := strings.Repeat("a", MaxFieldBytes+1)
	csv := "name,url,username,password,note\n" +
		"A,https://a.example,a,pass,\n" +
		",https://noTitle.example,x,pass,\n" +
		"B," + oversized + ",b,pass,\n" +
		"C,https://c.example,c,pass,\n"

	result, err := Parse(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if result.InvalidRows != 2 {
		t.Errorf("InvalidRows = %d, want 2", result.InvalidRows)
	}
	if len(result.Entries) != 2 {
		t.Fatalf("len(Entries) = %d, want 2", len(result.Entries))
	}
	if result.Entries[0].Title != "A" || result.Entries[1].Title != "C" {
		t.Errorf("Entries = %+v", result.Entries)
	}
}
