package googlecsv

import (
	"reflect"
	"testing"
)

// TestImportReportFieldsAreAllCounts enforces spec section 13.9
// structurally: the report must never contain passwords, usernames, URLs,
// titles, or notes. Every field being an int means there is nowhere to put
// a string secret even by accident — this test fails loudly if a future
// edit adds one.
func TestImportReportFieldsAreAllCounts(t *testing.T) {
	typ := reflect.TypeOf(ImportReport{})
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if f.Type.Kind() != reflect.Int {
			t.Errorf("field %s has kind %s, want int (spec section 13.9 forbids non-count fields)", f.Name, f.Type.Kind())
		}
	}
}

func TestImportReportStringFormat(t *testing.T) {
	r := ImportReport{
		Imported:               174,
		Skipped:                5,
		DuplicatesKept:         3,
		InvalidRows:            1,
		UnsupportedCredentials: 0,
	}
	want := "Imported: 174\nSkipped: 5\nDuplicates kept: 3\nInvalid rows: 1\nUnsupported credentials: 0"
	if got := r.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestImportReportZeroValueFormats(t *testing.T) {
	var r ImportReport
	want := "Imported: 0\nSkipped: 0\nDuplicates kept: 0\nInvalid rows: 0\nUnsupported credentials: 0"
	if got := r.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
