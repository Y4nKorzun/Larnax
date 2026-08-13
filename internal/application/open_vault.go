package application

import (
	"io"

	"github.com/Y4nKorzun/Larnax/internal/infrastructure/kdbx"
)

// OpenedVault is the result of opening a KDBX file: the document ready
// for editing, plus whether spec section 15.4's round-trip preservation
// gate forced it into read-only mode.
type OpenedVault struct {
	Document *kdbx.Document

	// ReadOnly is true when Unsupported is non-empty. Callers must refuse
	// SaveVault (a later piece of work) while this is true — writing back
	// would silently drop every construct named in Unsupported.
	ReadOnly bool

	// Unsupported lists the KDBX constructs found in the file that this
	// application's domain model cannot round-trip without data loss
	// (kdbx.DetectUnsupportedFeatures). Empty when ReadOnly is false.
	Unsupported []kdbx.UnsupportedFeature
}

// OpenVault opens the KDBX file read from r with masterPassphrase and
// decides whether it is safe to ever save back. Spec section 15.4: if the
// file contains a construct the application cannot safely preserve, it
// opens read-only rather than risk silently corrupting it on save.
func OpenVault(r io.Reader, masterPassphrase string) (*OpenedVault, error) {
	doc, err := kdbx.Open(r, masterPassphrase)
	if err != nil {
		return nil, err
	}

	unsupported := kdbx.DetectUnsupportedFeatures(doc.Database)
	return &OpenedVault{
		Document:    doc,
		ReadOnly:    len(unsupported) > 0,
		Unsupported: unsupported,
	}, nil
}
