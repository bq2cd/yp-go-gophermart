package domain

// PasswordPlain represents user's password in plaintext.
type PasswordPlain string

// PasswordHash represents user's password cryptographic hash as a slice of bytes.
type PasswordHash []byte

// IsEmpty returns [true] if password is an empty string.
func (p PasswordPlain) IsEmpty() bool {
	return len(p) == 0
}
