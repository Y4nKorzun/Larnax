package domain

import "errors"

var (
	ErrGroupNotFound        = errors.New("domain: group not found")
	ErrEntryNotFound        = errors.New("domain: entry not found")
	ErrGroupMustHaveParent  = errors.New("domain: only the vault's root group may have no parent")
	ErrRootMustHaveNoParent = errors.New("domain: a vault's root group must not have a parent")
	ErrCannotMoveRootGroup  = errors.New("domain: the root group cannot be moved")
	ErrCyclicGroupMove      = errors.New("domain: move would make a group its own ancestor")
	ErrSecretCleared        = errors.New("domain: secret has already been cleared")
)
