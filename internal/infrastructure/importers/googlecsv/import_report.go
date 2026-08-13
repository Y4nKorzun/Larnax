package googlecsv

import "fmt"

// ImportReport is the non-secret summary an import run can save (spec
// section 13.9). Every field is a plain count on purpose: the report must
// never contain passwords, usernames, URLs, titles, notes, or TOTP
// secrets, and an int has nowhere to put any of those even by accident.
type ImportReport struct {
	Imported               int
	Skipped                int
	DuplicatesKept         int
	InvalidRows            int
	UnsupportedCredentials int
}

// String matches spec section 13.9's example layout exactly.
func (r ImportReport) String() string {
	return fmt.Sprintf(
		"Imported: %d\nSkipped: %d\nDuplicates kept: %d\nInvalid rows: %d\nUnsupported credentials: %d",
		r.Imported, r.Skipped, r.DuplicatesKept, r.InvalidRows, r.UnsupportedCredentials,
	)
}
