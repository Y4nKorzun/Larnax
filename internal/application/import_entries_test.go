package application

import (
	"errors"
	"testing"

	"github.com/Y4nKorzun/Larnax/internal/domain"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/importers/googlecsv"
)

func importedGitHub() googlecsv.ImportedEntry {
	return googlecsv.ImportedEntry{
		Title:    "GitHub",
		Username: "octocat",
		Password: domain.NewSecretFromString("hunter2"),
		URL:      "https://github.com",
	}
}

func TestImportEntriesAddsNewEntries(t *testing.T) {
	v := domain.NewVault("test vault")
	stack := &CommandStack{}
	plan := ImportPlan{New: []googlecsv.ImportedEntry{importedGitHub()}}

	result, err := ImportEntries(v, stack, v.RootGroupID(), plan)
	if err != nil {
		t.Fatalf("ImportEntries() error = %v", err)
	}
	if result.Imported != 1 {
		t.Errorf("Imported = %d, want 1", result.Imported)
	}

	entries := v.AllEntries()
	if len(entries) != 1 {
		t.Fatalf("vault has %d entries, want 1", len(entries))
	}
	if entries[0].Title != "GitHub" || entries[0].Username != "octocat" {
		t.Errorf("entry = %+v, want Title=GitHub Username=octocat", entries[0])
	}
}

func TestImportEntriesRejectsUnknownParent(t *testing.T) {
	v := domain.NewVault("test vault")
	stack := &CommandStack{}
	plan := ImportPlan{New: []googlecsv.ImportedEntry{importedGitHub()}}

	if _, err := ImportEntries(v, stack, domain.NewGroupID(), plan); !errors.Is(err, domain.ErrGroupNotFound) {
		t.Errorf("ImportEntries() error = %v, want %v", err, domain.ErrGroupNotFound)
	}
	if len(v.AllEntries()) != 0 {
		t.Error("vault was mutated despite an unknown parent group")
	}
}

func existingVaultWithGitHub(t *testing.T) (*domain.Vault, domain.Entry) {
	t.Helper()
	v := domain.NewVault("test vault")
	existing := domain.NewEntry(v.RootGroupID(), "GitHub (old)")
	existing.Username = "octocat"
	existing.URL = "https://github.com"
	existing.Password = domain.NewSecretFromString("old-password")
	if err := v.AddEntry(existing); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}
	return v, existing
}

func TestImportEntriesKeepExistingDoesNotMutateVault(t *testing.T) {
	v, existing := existingVaultWithGitHub(t)
	stack := &CommandStack{}
	plan := ImportPlan{Resolved: []ResolvedDuplicate{{
		Match:      googlecsv.DuplicateMatch{Imported: importedGitHub(), Existing: existing},
		Resolution: KeepExisting,
	}}}

	result, err := ImportEntries(v, stack, v.RootGroupID(), plan)
	if err != nil {
		t.Fatalf("ImportEntries() error = %v", err)
	}
	if result.Skipped != 1 {
		t.Errorf("Skipped = %d, want 1", result.Skipped)
	}
	if len(v.AllEntries()) != 1 {
		t.Fatalf("vault has %d entries, want 1 (unchanged)", len(v.AllEntries()))
	}
	got, _ := v.Entry(existing.ID)
	if got.Title != "GitHub (old)" {
		t.Errorf("existing.Title = %q, want unchanged %q", got.Title, "GitHub (old)")
	}
}

func TestImportEntriesSkipDoesNotMutateVault(t *testing.T) {
	v, existing := existingVaultWithGitHub(t)
	stack := &CommandStack{}
	plan := ImportPlan{Resolved: []ResolvedDuplicate{{
		Match:      googlecsv.DuplicateMatch{Imported: importedGitHub(), Existing: existing},
		Resolution: Skip,
	}}}

	result, err := ImportEntries(v, stack, v.RootGroupID(), plan)
	if err != nil {
		t.Fatalf("ImportEntries() error = %v", err)
	}
	if result.Skipped != 1 {
		t.Errorf("Skipped = %d, want 1", result.Skipped)
	}
	if len(v.AllEntries()) != 1 {
		t.Errorf("vault has %d entries, want 1 (unchanged)", len(v.AllEntries()))
	}
}

func TestImportEntriesKeepBothAddsSecondEntry(t *testing.T) {
	v, existing := existingVaultWithGitHub(t)
	stack := &CommandStack{}
	plan := ImportPlan{Resolved: []ResolvedDuplicate{{
		Match:      googlecsv.DuplicateMatch{Imported: importedGitHub(), Existing: existing},
		Resolution: KeepBoth,
	}}}

	result, err := ImportEntries(v, stack, v.RootGroupID(), plan)
	if err != nil {
		t.Fatalf("ImportEntries() error = %v", err)
	}
	if result.Imported != 1 || result.DuplicatesKept != 1 {
		t.Errorf("result = %+v, want Imported=1 DuplicatesKept=1", result)
	}
	if len(v.AllEntries()) != 2 {
		t.Fatalf("vault has %d entries, want 2", len(v.AllEntries()))
	}
}

func TestImportEntriesUseImportedReplacesContentInPlace(t *testing.T) {
	v, existing := existingVaultWithGitHub(t)
	stack := &CommandStack{}
	plan := ImportPlan{Resolved: []ResolvedDuplicate{{
		Match:      googlecsv.DuplicateMatch{Imported: importedGitHub(), Existing: existing},
		Resolution: UseImported,
	}}}

	result, err := ImportEntries(v, stack, v.RootGroupID(), plan)
	if err != nil {
		t.Fatalf("ImportEntries() error = %v", err)
	}
	if result.Imported != 1 {
		t.Errorf("Imported = %d, want 1", result.Imported)
	}
	if len(v.AllEntries()) != 1 {
		t.Fatalf("vault has %d entries, want 1 (replaced in place)", len(v.AllEntries()))
	}

	got, err := v.Entry(existing.ID)
	if err != nil {
		t.Fatalf("Entry() error = %v (ID should be unchanged)", err)
	}
	if got.Title != "GitHub" {
		t.Errorf("Title = %q, want %q (imported value)", got.Title, "GitHub")
	}
	if !got.CreatedAt.Equal(existing.CreatedAt) {
		t.Errorf("CreatedAt = %v, want preserved %v", got.CreatedAt, existing.CreatedAt)
	}
}

func TestImportEntriesIsUndoableAsOneBatch(t *testing.T) {
	v := domain.NewVault("test vault")
	stack := &CommandStack{}
	plan := ImportPlan{New: []googlecsv.ImportedEntry{importedGitHub(), importedGitHub()}}

	if _, err := ImportEntries(v, stack, v.RootGroupID(), plan); err != nil {
		t.Fatalf("ImportEntries() error = %v", err)
	}
	if len(v.AllEntries()) != 2 {
		t.Fatalf("vault has %d entries, want 2", len(v.AllEntries()))
	}

	for i := 0; i < 2; i++ {
		ok, err := stack.Undo(v)
		if err != nil || !ok {
			t.Fatalf("Undo() #%d = (%v, %v), want (true, nil)", i, ok, err)
		}
	}
	if len(v.AllEntries()) != 0 {
		t.Errorf("vault has %d entries after undoing the whole import, want 0", len(v.AllEntries()))
	}
}
