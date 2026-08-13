package kdbx

import (
	"sort"

	"github.com/tobischo/gokeepasslib/v3"
)

// UnsupportedFeature names a KDBX construct this application cannot yet
// safely preserve through a decode-then-encode round trip. Detecting one
// is spec section 15.4's trigger for read-only fallback: "if the file
// contains a construct the application cannot safely save, it opens in
// read-only mode."
//
// This only covers constructs with no home in the current domain model
// (internal/domain's Entry/Group have no field for them at all) or that
// spec section 15.4's own preservation checklist names explicitly.
// Database-level settings — KDF parameters, cipher, compression — are
// deliberately not covered here: spec section 30's open question #4 asks
// which cipher/KDF combinations to even allow in write mode, a policy
// this application hasn't decided yet either, so a detector can't
// confidently flag combinations nobody has set a rule for.
//
// Entry.Tags is deliberately NOT flagged: domain.Entry already has a Tags
// field, so it is representable even though no mapper wires it up yet.
// Group-level tags are flagged (FeatureGroupTags) because domain.Group has
// no equivalent field at all.
type UnsupportedFeature string

const (
	FeatureAttachments         UnsupportedFeature = "attachments"
	FeatureEntryHistory        UnsupportedFeature = "entry-history"
	FeatureGroupTags           UnsupportedFeature = "group-tags"
	FeatureCustomData          UnsupportedFeature = "custom-data"
	FeatureCustomIcons         UnsupportedFeature = "custom-icons"
	FeaturePreviousParentGroup UnsupportedFeature = "previous-parent-group"
	FeatureDeletedObjects      UnsupportedFeature = "deleted-objects"
)

// DetectUnsupportedFeatures walks db's full group/entry tree and its
// metadata, returning the sorted, deduplicated set of unsupported
// features present. An empty result means db is safe to round-trip
// through this application's current domain model.
func DetectUnsupportedFeatures(db *gokeepasslib.Database) []UnsupportedFeature {
	found := make(map[UnsupportedFeature]bool)

	if db.Content == nil {
		return nil
	}

	if db.Content.Meta != nil {
		if len(db.Content.Meta.CustomIcons) > 0 {
			found[FeatureCustomIcons] = true
		}
		if len(db.Content.Meta.CustomData) > 0 {
			found[FeatureCustomData] = true
		}
	}

	if db.Content.Root != nil {
		if len(db.Content.Root.DeletedObjects) > 0 {
			found[FeatureDeletedObjects] = true
		}
		for _, g := range db.Content.Root.Groups {
			scanGroup(g, found)
		}
	}

	features := make([]UnsupportedFeature, 0, len(found))
	for f := range found {
		features = append(features, f)
	}
	sort.Slice(features, func(i, j int) bool { return features[i] < features[j] })
	return features
}

func scanGroup(g gokeepasslib.Group, found map[UnsupportedFeature]bool) {
	if g.Tags != "" {
		found[FeatureGroupTags] = true
	}
	if len(g.CustomData) > 0 {
		found[FeatureCustomData] = true
	}
	if g.PreviousParentGroup != nil {
		found[FeaturePreviousParentGroup] = true
	}

	for _, e := range g.Entries {
		scanEntry(e, found)
	}
	for _, child := range g.Groups {
		scanGroup(child, found)
	}
}

func scanEntry(e gokeepasslib.Entry, found map[UnsupportedFeature]bool) {
	if len(e.Binaries) > 0 {
		found[FeatureAttachments] = true
	}
	if len(e.Histories) > 0 {
		found[FeatureEntryHistory] = true
	}
	if len(e.CustomData) > 0 {
		found[FeatureCustomData] = true
	}
	if e.PreviousParentGroup != nil {
		found[FeaturePreviousParentGroup] = true
	}
}
