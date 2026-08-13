package googlecsv

import (
	"testing"

	"github.com/Y4nKorzun/Larnax/internal/domain"
)

func TestDetectDuplicatesFindsExactMatch(t *testing.T) {
	existing := []domain.Entry{
		{Title: "GitHub", Username: "user@example.com", URL: "https://github.com"},
	}
	imported := []ImportedEntry{
		{SourceRow: 2, Title: "GitHub", Username: "user@example.com", URL: "https://github.com"},
	}

	matches := DetectDuplicates(existing, imported)
	if len(matches) != 1 {
		t.Fatalf("len(matches) = %d, want 1", len(matches))
	}
	if matches[0].Existing.Title != "GitHub" || matches[0].Imported.SourceRow != 2 {
		t.Errorf("unexpected match: %+v", matches[0])
	}
	if !matches[0].TitleMatches {
		t.Error("TitleMatches = false, want true (identical titles)")
	}
}

func TestDetectDuplicatesTitleIsNotPartOfIdentity(t *testing.T) {
	existing := []domain.Entry{
		{Title: "GitHub Work", Username: "user@example.com", URL: "https://github.com"},
	}
	imported := []ImportedEntry{
		{Title: "GitHub Personal", Username: "user@example.com", URL: "https://github.com"},
	}

	matches := DetectDuplicates(existing, imported)
	if len(matches) != 1 {
		t.Fatalf("len(matches) = %d, want 1 (different titles must not prevent a match)", len(matches))
	}
	if matches[0].TitleMatches {
		t.Error("TitleMatches = true, want false (titles differ)")
	}
}

func TestDetectDuplicatesSameTitleDifferentSiteIsNotAMatch(t *testing.T) {
	existing := []domain.Entry{
		{Title: "GitHub", Username: "alice", URL: "https://github.com"},
	}
	imported := []ImportedEntry{
		{Title: "GitHub", Username: "alice", URL: "https://gitlab.com"},
	}

	matches := DetectDuplicates(existing, imported)
	if len(matches) != 0 {
		t.Errorf("len(matches) = %d, want 0 (identical titles for different sites is not a duplicate)", len(matches))
	}
}

func TestDetectDuplicatesNoMatchForDifferentUsername(t *testing.T) {
	existing := []domain.Entry{
		{Title: "GitHub", Username: "alice@example.com", URL: "https://github.com"},
	}
	imported := []ImportedEntry{
		{Title: "GitHub", Username: "bob@example.com", URL: "https://github.com"},
	}

	if matches := DetectDuplicates(existing, imported); len(matches) != 0 {
		t.Errorf("len(matches) = %d, want 0", len(matches))
	}
}

func TestDetectDuplicatesNoMatchForDifferentOrigin(t *testing.T) {
	existing := []domain.Entry{
		{Title: "GitHub", Username: "alice", URL: "https://github.com"},
	}
	imported := []ImportedEntry{
		{Title: "GitLab", Username: "alice", URL: "https://gitlab.com"},
	}

	if matches := DetectDuplicates(existing, imported); len(matches) != 0 {
		t.Errorf("len(matches) = %d, want 0", len(matches))
	}
}

func TestDetectDuplicatesHandlesMissingURLOrUsername(t *testing.T) {
	existing := []domain.Entry{
		{Title: "No URL", Username: "alice"},
		{Title: "No Username", URL: "https://github.com"},
	}
	imported := []ImportedEntry{
		{Title: "No URL", Username: "alice"},
		{Title: "No Username", URL: "https://github.com"},
	}

	if matches := DetectDuplicates(existing, imported); len(matches) != 0 {
		t.Errorf("len(matches) = %d, want 0 (entries missing URL or username must never match on emptiness alone)", len(matches))
	}
}

// TestDetectDuplicatesNormalizesBeforeComparing proves normalization is
// actually applied end-to-end by DetectDuplicates, not just unit-tested in
// isolation elsewhere.
func TestDetectDuplicatesNormalizesBeforeComparing(t *testing.T) {
	existing := []domain.Entry{
		{Title: "GitHub", Username: "User@Example.com", URL: "https://github.com"},
	}
	imported := []ImportedEntry{
		{Title: "GitHub", Username: "  user@example.com  ", URL: "www.github.com"},
	}

	if matches := DetectDuplicates(existing, imported); len(matches) != 1 {
		t.Errorf("len(matches) = %d, want 1 (case/whitespace/www/scheme differences should still match)", len(matches))
	}
}

func TestDetectDuplicatesNoFalsePositivesWhenNothingMatches(t *testing.T) {
	existing := []domain.Entry{
		{Title: "GitHub", Username: "alice", URL: "https://github.com"},
		{Title: "AWS", Username: "alice", URL: "https://aws.amazon.com"},
	}
	imported := []ImportedEntry{
		{Title: "Hetzner", Username: "alice", URL: "https://hetzner.com"},
		{Title: "GitLab", Username: "bob", URL: "https://gitlab.com"},
	}

	if matches := DetectDuplicates(existing, imported); len(matches) != 0 {
		t.Errorf("len(matches) = %d, want 0", len(matches))
	}
}

func TestDetectDuplicatesMatchesMultipleExistingEntries(t *testing.T) {
	existing := []domain.Entry{
		{Title: "GitHub old", Username: "alice", URL: "https://github.com"},
		{Title: "GitHub duplicate", Username: "alice", URL: "https://github.com"},
	}
	imported := []ImportedEntry{
		{Title: "GitHub", Username: "alice", URL: "https://github.com"},
	}

	matches := DetectDuplicates(existing, imported)
	if len(matches) != 2 {
		t.Fatalf("len(matches) = %d, want 2 (one per pre-existing duplicate)", len(matches))
	}
}
