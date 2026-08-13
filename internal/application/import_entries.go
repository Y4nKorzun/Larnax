package application

import (
	"github.com/Y4nKorzun/Larnax/internal/domain"
	"github.com/Y4nKorzun/Larnax/internal/infrastructure/importers/googlecsv"
)

// DuplicateResolution is the user's spec section 13.6 choice for one
// possible duplicate — the [k]/[i]/[b]/[s] prompt itself is a TUI concern
// ImportEntries never shows; it only acts on the decision already made.
type DuplicateResolution int

const (
	KeepExisting DuplicateResolution = iota // [k]: discard the imported entry
	UseImported                             // [i]: replace the existing entry's content
	KeepBoth                                // [b]: add the imported entry alongside the existing one
	Skip                                    // [s]: same effect as KeepExisting, distinct user intent
)

// ResolvedDuplicate pairs a detected duplicate with how the caller
// already decided to handle it.
type ResolvedDuplicate struct {
	Match      googlecsv.DuplicateMatch
	Resolution DuplicateResolution
}

// ImportPlan is spec section 13.7 step 5's "in-memory change set": every
// imported row, already sorted into entries with no duplicate (always
// added) and entries that matched one (each carrying its resolution).
// Building this from googlecsv.Parse's and googlecsv.DetectDuplicates'
// output — and collecting the per-duplicate decision — is a TUI concern;
// ImportEntries only applies an already-complete plan.
type ImportPlan struct {
	New      []googlecsv.ImportedEntry
	Resolved []ResolvedDuplicate
}

// ImportResult summarizes what ImportEntries did, in the same non-secret
// shape as googlecsv.ImportReport (spec section 13.9). The caller merges
// this with InvalidRows and UnsupportedCredentials, which come from steps
// ImportEntries never sees — parsing and passkey detection happen before
// a plan reaches here.
type ImportResult struct {
	Imported       int
	Skipped        int
	DuplicatesKept int
}

// ImportEntries applies plan to vault as one transaction on stack (spec
// section 13.7: the vault is not touched until every row's decision is
// already known). parent is the group new and kept-both entries are
// added under; an entry resolved as UseImported keeps its original
// group instead.
//
// ImportEntries validates that parent and every Resolved.Match.Existing
// still exist in vault *before* applying anything, so a failure leaves
// vault completely untouched (spec: "if any step before replacement
// fails, the original database stays unchanged"). This assumes plan was
// built from vault's own current state with nothing else mutating vault
// in between — true for the single synchronous call spec's import wizard
// makes, and not a condition ImportEntries can otherwise detect.
func ImportEntries(vault *domain.Vault, stack *CommandStack, parent domain.GroupID, plan ImportPlan) (ImportResult, error) {
	if _, err := vault.Group(parent); err != nil {
		return ImportResult{}, err
	}
	for _, resolved := range plan.Resolved {
		if resolved.Resolution == UseImported {
			if _, err := vault.Entry(resolved.Match.Existing.ID); err != nil {
				return ImportResult{}, err
			}
		}
	}

	var result ImportResult

	for _, imported := range plan.New {
		entry := entryFromImport(parent, imported)
		if err := stack.Do(vault, NewAddEntryCommand(entry)); err != nil {
			return ImportResult{}, err
		}
		result.Imported++
	}

	for _, resolved := range plan.Resolved {
		switch resolved.Resolution {
		case KeepBoth:
			entry := entryFromImport(parent, resolved.Match.Imported)
			if err := stack.Do(vault, NewAddEntryCommand(entry)); err != nil {
				return ImportResult{}, err
			}
			result.Imported++
			result.DuplicatesKept++

		case UseImported:
			before := resolved.Match.Existing
			after := entryFromImport(before.ParentGroup, resolved.Match.Imported)
			after.ID = before.ID
			after.CreatedAt = before.CreatedAt
			if err := stack.Do(vault, NewUpdateEntryCommand(before, after)); err != nil {
				return ImportResult{}, err
			}
			result.Imported++

		case KeepExisting, Skip:
			result.Skipped++
		}
	}

	return result, nil
}

func entryFromImport(parent domain.GroupID, imported googlecsv.ImportedEntry) domain.Entry {
	entry := domain.NewEntry(parent, imported.Title)
	entry.Username = imported.Username
	entry.Password = imported.Password
	entry.URL = imported.URL
	entry.Notes = imported.Notes
	entry.Tags = imported.Tags
	return entry
}
