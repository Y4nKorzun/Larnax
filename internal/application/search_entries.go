package application

import (
	"sort"
	"strings"

	"github.com/Y4nKorzun/Larnax/internal/domain"
)

// SearchableEntry is what the functions in this file examine: exactly the
// fields spec section 12.1 permits searching by default (Title, Username,
// URL, Tags, and the entry's resolved group path), plus an ID so callers
// can map results back to a domain.Entry. Password, TOTP secret,
// attachment content, and entry history are deliberately absent — there
// is nothing here to search, because they were never put in this struct.
//
// GroupPath is resolved by the caller (e.g. by walking domain.Vault's
// group tree) rather than by this file, so search stays testable without
// a live Vault.
type SearchableEntry struct {
	ID        domain.EntryID
	Title     string
	Username  string
	URL       string
	Tags      []string
	GroupPath string
}

// Search implements spec section 12.1's normal "/" search: a
// case-insensitive substring match across Title, Username, URL, Tags, and
// GroupPath. An empty query returns every entry unfiltered, matching how
// a "type to filter" search box behaves before any input.
func Search(entries []SearchableEntry, query string) []SearchableEntry {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return append([]SearchableEntry(nil), entries...)
	}

	var results []SearchableEntry
	for _, e := range entries {
		if matchesQuery(e, query) {
			results = append(results, e)
		}
	}
	return results
}

func matchesQuery(e SearchableEntry, query string) bool {
	if strings.Contains(strings.ToLower(e.Title), query) ||
		strings.Contains(strings.ToLower(e.Username), query) ||
		strings.Contains(strings.ToLower(e.URL), query) ||
		strings.Contains(strings.ToLower(e.GroupPath), query) {
		return true
	}
	for _, tag := range e.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}

// FuzzyResult pairs a SearchableEntry with the score FuzzyFind ranked it
// by.
type FuzzyResult struct {
	Entry SearchableEntry
	Score int
}

// FuzzyFind implements spec section 12.2's fuzzy finder. An entry is a
// candidate if it fuzzy-matches (as a subsequence) any of the same fields
// spec section 12.1 makes searchable; results are sorted by:
//  1. exact title match;
//  2. title prefix match;
//  3. best fuzzy score across all searchable fields;
//  4. recent usage — deliberately not implemented: spec section 12.2
//     says this must not be persisted between runs until the privacy
//     question around it is resolved, so there is nothing to rank by
//     here yet;
//  5. a stable tie-breaker — sort.SliceStable preserves input order for
//     anything the rules above don't distinguish.
//
// An empty query returns every entry, unscored, in original order.
func FuzzyFind(entries []SearchableEntry, query string) []FuzzyResult {
	query = strings.TrimSpace(query)
	if query == "" {
		results := make([]FuzzyResult, len(entries))
		for i, e := range entries {
			results[i] = FuzzyResult{Entry: e}
		}
		return results
	}

	lowerQuery := strings.ToLower(query)

	var results []FuzzyResult
	for _, e := range entries {
		score, ok := bestFuzzyScore(e, query)
		if !ok {
			continue
		}
		results = append(results, FuzzyResult{Entry: e, Score: score})
	}

	sort.SliceStable(results, func(i, j int) bool {
		a, b := results[i], results[j]

		aExact := strings.EqualFold(a.Entry.Title, query)
		bExact := strings.EqualFold(b.Entry.Title, query)
		if aExact != bExact {
			return aExact
		}

		aPrefix := strings.HasPrefix(strings.ToLower(a.Entry.Title), lowerQuery)
		bPrefix := strings.HasPrefix(strings.ToLower(b.Entry.Title), lowerQuery)
		if aPrefix != bPrefix {
			return aPrefix
		}

		return a.Score > b.Score
	})

	return results
}

// bestFuzzyScore returns the highest fuzzyMatch score for query across all
// of e's searchable fields, and whether any field matched at all.
func bestFuzzyScore(e SearchableEntry, query string) (score int, matched bool) {
	fields := make([]string, 0, 4+len(e.Tags))
	fields = append(fields, e.Title, e.Username, e.URL, e.GroupPath)
	fields = append(fields, e.Tags...)

	best := 0
	found := false
	for _, f := range fields {
		if s, ok := fuzzyMatch(f, query); ok && (!found || s > best) {
			best = s
			found = true
		}
	}
	return best, found
}

// fuzzyMatch reports whether every rune in query appears in target in
// order (not necessarily contiguous), case-insensitively — the standard
// subsequence fuzzy-match algorithm used by fzf, Telescope, and similar
// finders. When it matches, score rewards each matched character and
// penalizes gaps between consecutive matches, so a tighter, more
// contiguous match scores higher than a scattered one.
func fuzzyMatch(target, query string) (score int, ok bool) {
	targetRunes := []rune(strings.ToLower(target))
	queryRunes := []rune(strings.ToLower(query))
	if len(queryRunes) == 0 {
		return 0, true
	}

	qi := 0
	lastMatch := -1
	for ti, r := range targetRunes {
		if qi >= len(queryRunes) {
			break
		}
		if r == queryRunes[qi] {
			if lastMatch >= 0 {
				score -= ti - lastMatch - 1
			}
			score += 10
			lastMatch = ti
			qi++
		}
	}
	if qi < len(queryRunes) {
		return 0, false
	}
	return score, true
}
