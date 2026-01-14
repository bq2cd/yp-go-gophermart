package domain

const (
	// TokenEmptyValue represents an empty value for [Token].
	TokenEmptyValue Token = ""
)

// Token represents a JWT token issued for an authenticated user.
type Token string

// String returns string representation of the [Token].
func (t Token) String() string {
	return string(t)
}
