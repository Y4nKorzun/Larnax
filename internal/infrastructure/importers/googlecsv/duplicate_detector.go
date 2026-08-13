package googlecsv

import (
	"strings"

	"github.com/Y4nKorzun/Larnax/internal/domain"
)

// ImportedEntry is the intermediate representation of one parsed CSV row
// (spec section 13.4), before it becomes a domain.Entry. CSV parsing itself
// (parser.go, header_mapping.go) is separate, later work; this type exists
// here because DetectDuplicates needs a concrete shape to compare against.
type ImportedEntry struct {
	SourceRow int
	Title     string
	Username  string
	Password  domain.Secret
	URL       string
	Notes     string
	Tags      []string
}

// DuplicateMatch pairs an imported entry with the existing vault entry it
// possibly duplicates (spec section 13.6).
type DuplicateMatch struct {
	Imported     ImportedEntry
	Existing     domain.Entry
	TitleMatches bool
}

// DetectDuplicates flags each imported entry whose normalized origin +
// normalized username matches an existing vault entry — spec section
// 13.6's duplicate key. Title is surfaced as an additional signal via
// DuplicateMatch.TitleMatches, but never participates in the identity key
// itself: two entries with different titles for the same site and username
// are still a duplicate, and two entries that happen to share a title for
// different sites are not.
//
// An entry on either side with no comparable URL or username is never
// matched — treating a missing identity as equal to another missing
// identity would flood the result with false positives instead of
// correctly reporting "not enough information to compare".
func DetectDuplicates(existing []domain.Entry, imported []ImportedEntry) []DuplicateMatch {
	byKey := make(map[string][]domain.Entry, len(existing))
	for _, e := range existing {
		key := duplicateKey(e.URL, e.Username)
		if key == "" {
			continue
		}
		byKey[key] = append(byKey[key], e)
	}

	var matches []DuplicateMatch
	for _, imp := range imported {
		key := duplicateKey(imp.URL, imp.Username)
		if key == "" {
			continue
		}
		for _, e := range byKey[key] {
			matches = append(matches, DuplicateMatch{
				Imported:     imp,
				Existing:     e,
				TitleMatches: strings.EqualFold(strings.TrimSpace(e.Title), strings.TrimSpace(imp.Title)),
			})
		}
	}
	return matches
}

func duplicateKey(rawURL, username string) string {
	origin := normalizeOrigin(rawURL)
	user := normalizeUsername(username)
	if origin == "" || user == "" {
		return ""
	}
	return origin + "|" + user
}
