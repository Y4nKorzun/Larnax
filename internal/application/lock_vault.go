package application

import "github.com/Y4nKorzun/Larnax/internal/domain"

// LockVault best-effort clears every secret this process holds for vault
// and stack: spec section 17.2 steps 3 and 4, "delete the decrypted
// domain model" and "clear temporary editor buffers."
//
// It deliberately does not cover the rest of spec section 17.2's lock
// sequence:
//   - step 1 (offer to save unsaved changes) and step 8 (ask for the
//     master passphrase again) are user-facing prompts the TUI layer owns;
//   - step 2 (close file handles) belongs to whoever opened the file —
//     this package's OpenVault/SaveVault take an io.Reader/io.Writer, they
//     never hold one themselves;
//   - step 6 (clear the clipboard, but only if it still holds our secret)
//     needs the ownership hash clipboard.SecureCopy returned for the last
//     copy, which only the copy_field call site has — LockVault has no way
//     to reconstruct it.
func LockVault(vault *domain.Vault, stack *CommandStack) {
	for _, entry := range vault.AllEntries() {
		clearEntrySecret(entry)
	}
	stack.Clear()
}
