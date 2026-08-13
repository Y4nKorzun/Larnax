package application

import (
	"errors"
	"testing"

	"github.com/Y4nKorzun/Larnax/internal/domain"
)

func TestLockVaultClearsEntryPasswords(t *testing.T) {
	v := domain.NewVault("test vault")
	entry := domain.NewEntry(v.RootGroupID(), "GitHub")
	entry.Password = domain.NewSecretFromString("hunter2")
	if err := v.AddEntry(entry); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}

	LockVault(v, &CommandStack{})

	got, err := v.Entry(entry.ID)
	if err != nil {
		t.Fatalf("Entry() error = %v", err)
	}
	if err := got.Password.Reveal(func([]byte) error { return nil }); !errors.Is(err, domain.ErrSecretCleared) {
		t.Errorf("Reveal() after LockVault error = %v, want %v", err, domain.ErrSecretCleared)
	}
}

func TestLockVaultClearsCustomFieldSecrets(t *testing.T) {
	v := domain.NewVault("test vault")
	entry := domain.NewEntry(v.RootGroupID(), "GitHub")
	entry.CustomFields = []domain.Field{
		{Name: "Recovery Code", Value: domain.NewSecretFromString("abc123")},
	}
	if err := v.AddEntry(entry); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}

	LockVault(v, &CommandStack{})

	got, err := v.Entry(entry.ID)
	if err != nil {
		t.Fatalf("Entry() error = %v", err)
	}
	err = got.CustomFields[0].Value.Reveal(func([]byte) error { return nil })
	if !errors.Is(err, domain.ErrSecretCleared) {
		t.Errorf("custom field Reveal() after LockVault error = %v, want %v", err, domain.ErrSecretCleared)
	}
}

func TestLockVaultClearsCommandStack(t *testing.T) {
	v := domain.NewVault("test vault")
	entry := domain.NewEntry(v.RootGroupID(), "GitHub")
	stack := &CommandStack{}
	if err := stack.Do(v, NewAddEntryCommand(entry)); err != nil {
		t.Fatalf("Do() error = %v", err)
	}

	LockVault(v, stack)

	if ok, err := stack.Undo(v); ok || err != nil {
		t.Errorf("Undo() after LockVault = (%v, %v), want (false, nil)", ok, err)
	}
}

func TestLockVaultHandlesEmptyVault(t *testing.T) {
	v := domain.NewVault("test vault")
	LockVault(v, &CommandStack{}) // must not panic
}
