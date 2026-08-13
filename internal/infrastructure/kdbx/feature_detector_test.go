package kdbx

import (
	"reflect"
	"testing"

	"github.com/tobischo/gokeepasslib/v3"
	w "github.com/tobischo/gokeepasslib/v3/wrappers"
)

func newTestDatabase() *gokeepasslib.Database {
	db := gokeepasslib.NewDatabase(gokeepasslib.WithDatabaseKDBXVersion41())
	db.Credentials = gokeepasslib.NewPasswordCredentials("test")
	return db
}

func TestDetectUnsupportedFeaturesCleanDatabaseIsEmpty(t *testing.T) {
	db := newTestDatabase()
	rootGroup := &db.Content.Root.Groups[0]
	rootGroup.Entries = nil

	child := gokeepasslib.NewGroup()
	child.Name = "Personal"
	entry := gokeepasslib.NewEntry()
	entry.Values = append(entry.Values, gokeepasslib.ValueData{
		Key:   "Title",
		Value: gokeepasslib.V{Content: "GitHub"},
	})
	entry.Tags = "work,dev" // entry-level tags ARE representable; must not be flagged
	child.Entries = append(child.Entries, entry)
	rootGroup.Groups = append(rootGroup.Groups, child)

	got := DetectUnsupportedFeatures(db)
	if len(got) != 0 {
		t.Errorf("DetectUnsupportedFeatures() = %v, want empty for a database using only supported fields", got)
	}
}

func TestDetectUnsupportedFeaturesHandlesNilContent(t *testing.T) {
	db := &gokeepasslib.Database{}
	got := DetectUnsupportedFeatures(db)
	if len(got) != 0 {
		t.Errorf("DetectUnsupportedFeatures() = %v, want empty for a database with nil Content", got)
	}
}

func TestDetectUnsupportedFeaturesFindsAttachments(t *testing.T) {
	db := newTestDatabase()
	rootGroup := &db.Content.Root.Groups[0]
	rootGroup.Entries[0].Binaries = []gokeepasslib.BinaryReference{{}}

	got := DetectUnsupportedFeatures(db)
	assertContains(t, got, FeatureAttachments)
}

func TestDetectUnsupportedFeaturesFindsEntryHistory(t *testing.T) {
	db := newTestDatabase()
	rootGroup := &db.Content.Root.Groups[0]
	rootGroup.Entries[0].Histories = []gokeepasslib.History{{}}

	got := DetectUnsupportedFeatures(db)
	assertContains(t, got, FeatureEntryHistory)
}

func TestDetectUnsupportedFeaturesFindsGroupTags(t *testing.T) {
	db := newTestDatabase()
	rootGroup := &db.Content.Root.Groups[0]
	rootGroup.Tags = "important"

	got := DetectUnsupportedFeatures(db)
	assertContains(t, got, FeatureGroupTags)
}

func TestDetectUnsupportedFeaturesFindsCustomDataAtEachLevel(t *testing.T) {
	cd := []gokeepasslib.CustomData{{Key: "k", Value: "v"}}

	t.Run("group", func(t *testing.T) {
		db := newTestDatabase()
		db.Content.Root.Groups[0].CustomData = cd
		assertContains(t, DetectUnsupportedFeatures(db), FeatureCustomData)
	})

	t.Run("entry", func(t *testing.T) {
		db := newTestDatabase()
		db.Content.Root.Groups[0].Entries[0].CustomData = cd
		assertContains(t, DetectUnsupportedFeatures(db), FeatureCustomData)
	})

	t.Run("meta", func(t *testing.T) {
		db := newTestDatabase()
		db.Content.Meta.CustomData = cd
		assertContains(t, DetectUnsupportedFeatures(db), FeatureCustomData)
	})
}

func TestDetectUnsupportedFeaturesFindsCustomIcons(t *testing.T) {
	db := newTestDatabase()
	db.Content.Meta.CustomIcons = []gokeepasslib.CustomIcon{{Data: "base64data"}}

	got := DetectUnsupportedFeatures(db)
	assertContains(t, got, FeatureCustomIcons)
}

func TestDetectUnsupportedFeaturesFindsPreviousParentGroup(t *testing.T) {
	uuid := gokeepasslib.NewUUID()

	t.Run("entry", func(t *testing.T) {
		db := newTestDatabase()
		db.Content.Root.Groups[0].Entries[0].PreviousParentGroup = &uuid
		assertContains(t, DetectUnsupportedFeatures(db), FeaturePreviousParentGroup)
	})

	t.Run("group", func(t *testing.T) {
		db := newTestDatabase()
		db.Content.Root.Groups[0].PreviousParentGroup = &uuid
		assertContains(t, DetectUnsupportedFeatures(db), FeaturePreviousParentGroup)
	})
}

func TestDetectUnsupportedFeaturesFindsDeletedObjects(t *testing.T) {
	db := newTestDatabase()
	db.Content.Root.DeletedObjects = []gokeepasslib.DeletedObjectData{
		{UUID: gokeepasslib.NewUUID(), DeletionTime: &w.TimeWrapper{}},
	}

	got := DetectUnsupportedFeatures(db)
	assertContains(t, got, FeatureDeletedObjects)
}

func TestDetectUnsupportedFeaturesRecursesIntoNestedGroups(t *testing.T) {
	db := newTestDatabase()
	rootGroup := &db.Content.Root.Groups[0]

	level1 := gokeepasslib.NewGroup()
	level1.Name = "Level1"
	level2 := gokeepasslib.NewGroup()
	level2.Name = "Level2"
	level2.Tags = "deep-tag" // buried two levels down

	level1.Groups = append(level1.Groups, level2)
	rootGroup.Groups = append(rootGroup.Groups, level1)

	got := DetectUnsupportedFeatures(db)
	assertContains(t, got, FeatureGroupTags)
}

func TestDetectUnsupportedFeaturesEntryTagsAreNotFlagged(t *testing.T) {
	db := newTestDatabase()
	db.Content.Root.Groups[0].Entries[0].Tags = "personal,important"

	got := DetectUnsupportedFeatures(db)
	if len(got) != 0 {
		t.Errorf("DetectUnsupportedFeatures() = %v, want empty (entry-level tags are representable via domain.Entry.Tags)", got)
	}
}

func TestDetectUnsupportedFeaturesResultIsSortedAndDeduplicated(t *testing.T) {
	db := newTestDatabase()
	rootGroup := &db.Content.Root.Groups[0]

	// Trigger FeatureAttachments from two different entries.
	rootGroup.Entries[0].Binaries = []gokeepasslib.BinaryReference{{}}
	second := gokeepasslib.NewEntry()
	second.Binaries = []gokeepasslib.BinaryReference{{}}
	rootGroup.Entries = append(rootGroup.Entries, second)
	// And a second, distinct feature.
	rootGroup.Tags = "tag"

	got := DetectUnsupportedFeatures(db)
	if len(got) != 2 {
		t.Fatalf("DetectUnsupportedFeatures() = %v, want exactly 2 deduplicated entries", got)
	}
	for i := 1; i < len(got); i++ {
		if got[i-1] >= got[i] {
			t.Errorf("result %v is not sorted (or has a duplicate) at index %d", got, i)
		}
	}

	// Determinism: calling again must yield the identical slice.
	again := DetectUnsupportedFeatures(db)
	if !reflect.DeepEqual(got, again) {
		t.Errorf("DetectUnsupportedFeatures() is not deterministic: %v vs %v", got, again)
	}
}

func assertContains(t *testing.T, features []UnsupportedFeature, want UnsupportedFeature) {
	t.Helper()
	for _, f := range features {
		if f == want {
			return
		}
	}
	t.Errorf("features = %v, want it to contain %q", features, want)
}
