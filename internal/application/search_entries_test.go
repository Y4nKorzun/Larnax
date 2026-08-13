package application

import "testing"

func TestSearchMatchesEachField(t *testing.T) {
	entries := []SearchableEntry{
		{Title: "GitHub", Username: "octocat", URL: "https://github.com", Tags: []string{"dev"}, GroupPath: "Personal/Work"},
	}

	cases := []string{"github", "octocat", "github.com", "dev", "personal"}
	for _, q := range cases {
		t.Run(q, func(t *testing.T) {
			if got := Search(entries, q); len(got) != 1 {
				t.Errorf("Search(%q) = %d results, want 1", q, len(got))
			}
		})
	}
}

func TestSearchIsCaseInsensitive(t *testing.T) {
	entries := []SearchableEntry{{Title: "GitHub"}}
	if got := Search(entries, "GITHUB"); len(got) != 1 {
		t.Errorf("Search() = %d results, want 1", len(got))
	}
}

func TestSearchEmptyQueryReturnsAllEntries(t *testing.T) {
	entries := []SearchableEntry{{Title: "A"}, {Title: "B"}}
	got := Search(entries, "")
	if len(got) != 2 {
		t.Errorf("Search(\"\") = %d results, want 2 (unfiltered)", len(got))
	}
}

func TestSearchNoMatchReturnsEmpty(t *testing.T) {
	entries := []SearchableEntry{{Title: "GitHub"}}
	if got := Search(entries, "gitlab"); len(got) != 0 {
		t.Errorf("Search() = %d results, want 0", len(got))
	}
}

func TestFuzzyFindExactTitleRanksFirst(t *testing.T) {
	entries := []SearchableEntry{
		{Title: "Old GitHub Archive"}, // fuzzy-matches "github" but not exact/prefix
		{Title: "GitHub"},             // exact match
		{Title: "GitHub Enterprise"},  // prefix match
	}

	results := FuzzyFind(entries, "GitHub")
	if len(results) != 3 {
		t.Fatalf("len(results) = %d, want 3", len(results))
	}
	if results[0].Entry.Title != "GitHub" {
		t.Errorf("results[0].Title = %q, want %q (exact match first)", results[0].Entry.Title, "GitHub")
	}
}

func TestFuzzyFindPrefixRanksAboveNonPrefixFuzzy(t *testing.T) {
	entries := []SearchableEntry{
		{Title: "Old GitHub Archive"}, // "github" is a fuzzy subsequence but not a prefix
		{Title: "GitHub Enterprise"},  // prefix match
	}

	results := FuzzyFind(entries, "github")
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	if results[0].Entry.Title != "GitHub Enterprise" {
		t.Errorf("results[0].Title = %q, want %q (prefix ranks above scattered fuzzy match)", results[0].Entry.Title, "GitHub Enterprise")
	}
}

func TestFuzzyFindTighterMatchScoresHigher(t *testing.T) {
	// Both are subsequence matches for "gh", neither a prefix or exact
	// match, so they're compared purely on fuzzy score.
	tight, ok := fuzzyMatch("xGHx", "gh")
	if !ok {
		t.Fatal("fuzzyMatch(xGHx, gh) did not match")
	}
	loose, ok := fuzzyMatch("xGxxxxHx", "gh")
	if !ok {
		t.Fatal("fuzzyMatch(xGxxxxHx, gh) did not match")
	}
	if tight <= loose {
		t.Errorf("tight match score (%d) should be greater than loose match score (%d)", tight, loose)
	}

	entries := []SearchableEntry{
		{Title: "xGxxxxHx"},
		{Title: "xGHx"},
	}
	results := FuzzyFind(entries, "gh")
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	if results[0].Entry.Title != "xGHx" {
		t.Errorf("results[0].Title = %q, want %q (tighter match first)", results[0].Entry.Title, "xGHx")
	}
}

func TestFuzzyFindMatchesAcrossFields(t *testing.T) {
	entries := []SearchableEntry{
		{Title: "Server Login", Username: "octocat"},
	}

	results := FuzzyFind(entries, "octocat")
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1 (title doesn't match but username does)", len(results))
	}
}

func TestFuzzyFindExcludesNonMatchingEntries(t *testing.T) {
	entries := []SearchableEntry{
		{Title: "GitHub", Username: "octocat", URL: "https://github.com"},
		{Title: "AWS Console", Username: "alice", URL: "https://aws.amazon.com"},
	}

	results := FuzzyFind(entries, "github")
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].Entry.Title != "GitHub" {
		t.Errorf("results[0].Title = %q, want %q", results[0].Entry.Title, "GitHub")
	}
}

func TestFuzzyFindStableTieBreakPreservesInputOrder(t *testing.T) {
	// Identical titles produce identical scores and identical
	// exact/prefix status, so nothing but input order can distinguish
	// them.
	entries := []SearchableEntry{
		{ID: [16]byte{1}, Title: "Duplicate"},
		{ID: [16]byte{2}, Title: "Duplicate"},
		{ID: [16]byte{3}, Title: "Duplicate"},
	}

	results := FuzzyFind(entries, "Duplicate")
	if len(results) != 3 {
		t.Fatalf("len(results) = %d, want 3", len(results))
	}
	for i, want := range []byte{1, 2, 3} {
		if results[i].Entry.ID[0] != want {
			t.Errorf("results[%d].ID[0] = %d, want %d (stable tie-break should preserve input order)", i, results[i].Entry.ID[0], want)
		}
	}
}

func TestFuzzyFindEmptyQueryReturnsAllInOriginalOrder(t *testing.T) {
	entries := []SearchableEntry{{Title: "B"}, {Title: "A"}, {Title: "C"}}
	results := FuzzyFind(entries, "")
	if len(results) != 3 {
		t.Fatalf("len(results) = %d, want 3", len(results))
	}
	want := []string{"B", "A", "C"}
	for i, w := range want {
		if results[i].Entry.Title != w {
			t.Errorf("results[%d].Title = %q, want %q (unfiltered, original order)", i, results[i].Entry.Title, w)
		}
	}
}

func TestFuzzyMatchHandlesUnicode(t *testing.T) {
	// Cyrillic input exercises the []rune-based implementation rather
	// than a byte-indexed one, which would misbehave on multi-byte UTF-8.
	score, ok := fuzzyMatch("Пароль от почты", "почт")
	if !ok {
		t.Fatal("fuzzyMatch() did not match a Cyrillic substring")
	}
	if score <= 0 {
		t.Errorf("score = %d, want positive for a contiguous match", score)
	}
}

func TestFuzzyMatchRejectsOutOfOrderCharacters(t *testing.T) {
	if _, ok := fuzzyMatch("GitHub", "hg"); ok {
		t.Error("fuzzyMatch(GitHub, hg) matched, want no match (h comes after g in the target)")
	}
}
