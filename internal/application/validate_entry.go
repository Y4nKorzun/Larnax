package application

import (
	"errors"
	"strings"

	"github.com/Y4nKorzun/Larnax/internal/domain"
)

// ErrEmptyTitle is returned by ValidateEntry when Title is empty or only
// whitespace. Every other standard field (spec section 9.1: Username,
// Password, URL, Notes) is optional — nothing else in this package or
// domain requires them, and inventing a stricter rule spec never states
// would just be arbitrary policy dressed up as validation.
var ErrEmptyTitle = errors.New("application: entry title must not be empty")

// ValidateEntry reports the first constraint violation found in entry,
// or nil if it is valid enough to add or update in a vault. This is a
// UI-facing check for the entry editor form (spec section 9.2), meant to
// run before AddEntry/UpdateEntry — those two don't validate content on
// their own, since domain.Vault's own invariants (spec 18.4) are about
// structural correctness (parent references, no cycles), not content
// quality.
func ValidateEntry(entry domain.Entry) error {
	if strings.TrimSpace(entry.Title) == "" {
		return ErrEmptyTitle
	}
	return nil
}
