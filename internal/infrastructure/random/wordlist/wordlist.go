// Package wordlist embeds the fixed, versioned wordlist spec section 7.4
// requires for passphrase generation. See README.md for provenance.
package wordlist

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed eff_large_wordlist.txt
var effLargeWordlistRaw string

// EFFLarge is the EFF Long Wordlist, parsed from the original
// dice-roll-indexed source file. It has exactly 7776 words with no
// duplicates (verified in wordlist_test.go).
var EFFLarge = mustParseWordlist(effLargeWordlistRaw)

func mustParseWordlist(raw string) []string {
	words, err := parseWordlist(raw)
	if err != nil {
		// The embedded file is a build-time constant, not user input — a
		// parse failure here means the embedded asset itself is corrupt,
		// which is a build-time bug, not a runtime condition to recover
		// from.
		panic("wordlist: " + err.Error())
	}
	return words
}

// parseWordlist extracts the word column from the EFF source format
// ("<5-digit dice roll>\t<word>" per line).
func parseWordlist(raw string) ([]string, error) {
	lines := strings.Split(strings.TrimRight(raw, "\n"), "\n")
	words := make([]string, 0, len(lines))
	for i, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 2 {
			return nil, fmt.Errorf("line %d: expected 2 tab-separated fields, got %d: %q", i+1, len(fields), line)
		}
		words = append(words, fields[1])
	}
	return words, nil
}
