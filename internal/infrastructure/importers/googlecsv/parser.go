package googlecsv

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/Y4nKorzun/Larnax/internal/domain"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/terminal"
)

// MaxFieldBytes caps a single CSV field's size — spec section 24.4 names
// "a CSV with millions of characters in one field" as a specific
// failure-injection scenario to defend against. A row with any field over
// this is counted as invalid rather than accepted with a truncated value.
// This check runs after encoding/csv has already parsed the row, so it
// bounds what gets accepted into the app's data model, not the transient
// memory used while csv.Reader parses one pathological field — a fully
// proactive guard would need a custom byte-limiting reader in front of
// csv.Reader, which is more machinery than this chunk's scope covers.
const MaxFieldBytes = 1 << 20 // 1 MiB

// headerAliases maps known Google Password Manager CSV column name
// variants to canonical fields. Spec section 13.4 is explicit that Google
// can rename or reorder columns, so this list is deliberately permissive
// rather than pinned to one exact export format this code can't
// independently verify against a live export from this environment.
var headerAliases = map[string]string{
	"name":           "title",
	"title":          "title",
	"url":            "url",
	"login_uri":      "url",
	"site":           "url",
	"username":       "username",
	"login_username": "username",
	"user":           "username",
	"password":       "password",
	"login_password": "password",
	"note":           "notes",
	"notes":          "notes",
}

var ErrNoHeader = errors.New("googlecsv: file has no header row")

// ParseResult is what parsing a Google CSV export produces. It carries
// enough information to build spec section 13.9's import report without
// this package needing to know about duplicate detection or the eventual
// transactional save.
type ParseResult struct {
	Entries     []ImportedEntry
	InvalidRows int
	// UnknownSchema is true when Title and Password could not both be
	// mapped from the header row — spec section 13.4's cue to open a
	// column-mapping wizard (a later, TUI-layer concern) instead of
	// guessing at column meaning.
	UnknownSchema bool
	// PasskeyCount is set when the header looks like a passkey export
	// (see DetectPasskeyHeader in passkeys.go) rather than a password
	// export — spec section 13.8's "unsupported credential types"
	// warning. Always 0 when UnknownSchema is false: a header that
	// mapped Title and Password successfully is a password export by
	// definition.
	PasskeyCount int
}

// Parse reads a Google Password Manager CSV export from r (spec sections
// 13.3-13.4). It never buffers the file to a temp location itself — the
// caller controls where r's bytes come from — and never logs or includes
// row content in an error, since spec section 13.3 forbids logging CSV
// rows. It does not interpret spreadsheet formulas: every field is read
// as a literal string, so a value like "=cmd|..." is stored and later
// compared/displayed as inert text, never evaluated.
//
// Title and Notes/Username/URL are passed through terminal.Sanitize
// before being stored, since they may reach a terminal at preview or
// browse time. Password is deliberately NOT sanitized here: it must
// survive import as the exact credential bytes, and sanitizing it would
// silently corrupt a real password that happens to contain an unusual
// byte. Password sanitization belongs at the point of *display* (the
// reveal feature, spec section 8.5), not at the point of storage.
func Parse(r io.Reader) (ParseResult, error) {
	bufReader := bufio.NewReader(r)
	stripBOM(bufReader)

	headerLine, err := bufReader.ReadString('\n')
	if err != nil && err != io.EOF {
		return ParseResult{}, fmt.Errorf("googlecsv: reading header: %w", err)
	}
	if strings.TrimSpace(headerLine) == "" {
		return ParseResult{}, ErrNoHeader
	}

	csvReader := csv.NewReader(io.MultiReader(strings.NewReader(headerLine), bufReader))
	csvReader.Comma = sniffDelimiter(headerLine)
	csvReader.LazyQuotes = true
	csvReader.FieldsPerRecord = -1 // tolerate ragged rows rather than aborting the whole file

	header, err := csvReader.Read()
	if err != nil {
		return ParseResult{}, fmt.Errorf("%w: %v", ErrNoHeader, err)
	}

	columnIndex, ok := mapHeader(header)
	if !ok {
		if DetectPasskeyHeader(header) {
			return ParseResult{UnknownSchema: true, PasskeyCount: countDataRows(csvReader)}, nil
		}
		return ParseResult{UnknownSchema: true}, nil
	}

	var result ParseResult
	sourceRow := 1 // the header itself is row 1
	for {
		sourceRow++
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.InvalidRows++
			continue
		}

		entry, ok := buildEntry(record, columnIndex, sourceRow)
		if !ok {
			result.InvalidRows++
			continue
		}
		result.Entries = append(result.Entries, entry)
	}

	return result, nil
}

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

func stripBOM(r *bufio.Reader) {
	peeked, err := r.Peek(len(utf8BOM))
	if err == nil && bytes.Equal(peeked, utf8BOM) {
		_, _ = r.Discard(len(utf8BOM))
	}
}

// sniffDelimiter picks whichever of comma, semicolon, or tab appears most
// often in headerLine, defaulting to comma on a tie (including the
// no-delimiter-characters-at-all case of a single-column header).
// Candidates are checked in a fixed order specifically so the result is
// deterministic — iterating a map here would make ties resolve randomly.
func sniffDelimiter(headerLine string) rune {
	candidates := []rune{',', ';', '\t'}
	best := ','
	bestCount := -1
	for _, d := range candidates {
		if count := strings.Count(headerLine, string(d)); count > bestCount {
			best = d
			bestCount = count
		}
	}
	return best
}

func mapHeader(header []string) (columnIndex map[string]int, ok bool) {
	columnIndex = make(map[string]int, len(header))
	for i, col := range header {
		if canonical, known := headerAliases[strings.ToLower(strings.TrimSpace(col))]; known {
			columnIndex[canonical] = i
		}
	}
	_, hasTitle := columnIndex["title"]
	_, hasPassword := columnIndex["password"]
	return columnIndex, hasTitle && hasPassword
}

func buildEntry(record []string, columnIndex map[string]int, sourceRow int) (ImportedEntry, bool) {
	for _, field := range record {
		if len(field) > MaxFieldBytes {
			return ImportedEntry{}, false
		}
	}

	get := func(key string) string {
		i, ok := columnIndex[key]
		if !ok || i >= len(record) {
			return ""
		}
		return record[i]
	}

	title := terminal.Sanitize(get("title"))
	if title == "" {
		return ImportedEntry{}, false
	}

	return ImportedEntry{
		SourceRow: sourceRow,
		Title:     title,
		Username:  terminal.Sanitize(get("username")),
		Password:  domain.NewSecretFromString(get("password")),
		URL:       terminal.Sanitize(get("url")),
		Notes:     terminal.Sanitize(get("notes")),
	}, true
}
