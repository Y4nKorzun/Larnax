package domain

// Field is a user-defined custom entry field beyond the standard
// Title/Username/Password/URL/Notes set (spec section 9.1). Its value is a
// Secret since KDBX allows custom fields to be marked protected, and there
// is no way to know in advance which custom fields a user will use for
// sensitive data.
type Field struct {
	Name  string
	Value Secret
}
