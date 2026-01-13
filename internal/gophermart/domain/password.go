package domain

// PasswordPlain represents user's password in plaintext.
type PasswordPlain string

// PasswordHash represents user's password cryptographic hash as a slice of bytes.
type PasswordHash []byte
