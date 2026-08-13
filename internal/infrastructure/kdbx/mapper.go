// Package kdbx bridges internal/domain's storage-agnostic types and
// gokeepasslib/v3's KDBX-shaped types (spec section 18.3). mapper.go holds
// the leaf conversions for a single Entry; decoder.go and encoder.go (a
// later piece of work) walk gokeepasslib's nested group tree and
// internal/domain.Vault's flat map using these conversions at each node.
package kdbx

import (
	"strings"
	"time"

	"github.com/tobischo/gokeepasslib/v3"
	w "github.com/tobischo/gokeepasslib/v3/wrappers"

	"github.com/Y4nKorzun/Larnax/internal/domain"
)

// Standard KDBX entry field keys (KDBX XML schema, not spec-defined —
// these are the exact strings the reference client and gokeepasslib both
// use as ValueData.Key).
const (
	fieldTitle    = "Title"
	fieldUserName = "UserName"
	fieldPassword = "Password"
	fieldURL      = "URL"
	fieldNotes    = "Notes"
)

// tagSeparator matches how the reference client stores an entry's tags:
// one Tags string, semicolon-separated, not a repeated XML element.
const tagSeparator = ";"

// standardFields is the set of ValueData keys entryToGKP always writes
// itself. entryFromGKP uses it to tell a standard field apart from a
// custom one when walking ge.Values back into domain.Entry.CustomFields.
var standardFields = map[string]bool{
	fieldTitle:    true,
	fieldUserName: true,
	fieldPassword: true,
	fieldURL:      true,
	fieldNotes:    true,
}

// valueData builds a gokeepasslib ValueData. protected marks it for the
// KDBX inner-stream encryption spec section 15.2 requires for secrets —
// callers pass true for Password and every custom field, since
// domain.Field has no way to say a given custom value is safe to leave
// unprotected (see internal/domain/field.go).
func valueData(key, content string, protected bool) gokeepasslib.ValueData {
	v := gokeepasslib.ValueData{Key: key, Value: gokeepasslib.V{Content: content}}
	if protected {
		v.Value.Protected = w.NewBoolWrapper(true)
	}
	return v
}

// revealString copies a Secret's bytes into a Go string so it can be
// handed to gokeepasslib's string-based Value.Content. The copy is a
// necessary consequence of gokeepasslib's API: unlike domain.Secret, a Go
// string cannot be zeroed on Clear (spec section 5.3's memory model is
// already honest that Go offers no stronger guarantee than best effort).
func revealString(s domain.Secret) (string, error) {
	var content string
	err := s.Reveal(func(value []byte) error {
		content = string(value)
		return nil
	})
	return content, err
}

func timeWrapper(t time.Time) *w.TimeWrapper {
	tw := w.TimeWrapper{Time: t}
	return &tw
}

// timeValue returns the zero time.Time for a nil wrapper — TimeData's
// pointer fields (see gokeepasslib's time_data.go) are only nil for
// content this package itself never produces, but decoding a foreign KDBX
// file (spec section 15.1) must still not panic on one.
func timeValue(tw *w.TimeWrapper) time.Time {
	if tw == nil {
		return time.Time{}
	}
	return tw.Time
}

// entryToGKP converts a domain.Entry into a gokeepasslib.Entry ready to be
// placed into a Database's group tree. It does not itself encrypt
// protected values — LockProtectedEntries (called once, database-wide,
// during save) does that.
func entryToGKP(e domain.Entry) (gokeepasslib.Entry, error) {
	password, err := revealString(e.Password)
	if err != nil {
		return gokeepasslib.Entry{}, err
	}

	ge := gokeepasslib.Entry{
		UUID: gokeepasslib.UUID(e.ID),
		Times: gokeepasslib.TimeData{
			CreationTime:         timeWrapper(e.CreatedAt),
			LastModificationTime: timeWrapper(e.ModifiedAt),
			LastAccessTime:       timeWrapper(e.ModifiedAt),
		},
		Tags: strings.Join(e.Tags, tagSeparator),
	}
	if e.ExpiresAt != nil {
		ge.Times.Expires = w.NewBoolWrapper(true)
		ge.Times.ExpiryTime = timeWrapper(*e.ExpiresAt)
	}

	ge.Values = append(ge.Values,
		valueData(fieldTitle, e.Title, false),
		valueData(fieldUserName, e.Username, false),
		valueData(fieldPassword, password, true),
		valueData(fieldURL, e.URL, false),
		valueData(fieldNotes, e.Notes, false),
	)

	for _, f := range e.CustomFields {
		content, err := revealString(f.Value)
		if err != nil {
			return gokeepasslib.Entry{}, err
		}
		ge.Values = append(ge.Values, valueData(f.Name, content, true))
	}

	return ge, nil
}

// entryFromGKP converts a decoded, already-unlocked gokeepasslib.Entry
// into a domain.Entry. It leaves ParentGroup unset — decoder.go (a later
// piece of work) fills that in from the entry's position in the source
// tree, which this function never sees.
func entryFromGKP(ge gokeepasslib.Entry) domain.Entry {
	e := domain.Entry{
		ID:         domain.EntryID(ge.UUID),
		Title:      ge.GetContent(fieldTitle),
		Username:   ge.GetContent(fieldUserName),
		Password:   domain.NewSecretFromString(ge.GetContent(fieldPassword)),
		URL:        ge.GetContent(fieldURL),
		Notes:      ge.GetContent(fieldNotes),
		Tags:       splitTags(ge.Tags),
		CreatedAt:  timeValue(ge.Times.CreationTime),
		ModifiedAt: timeValue(ge.Times.LastModificationTime),
	}
	if ge.Times.Expires.Bool && ge.Times.ExpiryTime != nil {
		t := ge.Times.ExpiryTime.Time
		e.ExpiresAt = &t
	}

	for _, v := range ge.Values {
		if standardFields[v.Key] {
			continue
		}
		e.CustomFields = append(e.CustomFields, domain.Field{
			Name:  v.Key,
			Value: domain.NewSecretFromString(v.Value.Content),
		})
	}

	return e
}

// splitTags reverses strings.Join(tags, tagSeparator), dropping empty
// segments so an entry with no tags decodes back to a nil slice rather
// than []string{""}.
func splitTags(tags string) []string {
	if tags == "" {
		return nil
	}
	var result []string
	for _, t := range strings.Split(tags, tagSeparator) {
		if t != "" {
			result = append(result, t)
		}
	}
	return result
}
