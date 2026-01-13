package domain

// Token represents a JWT token issued for an authenticated user.
type Token string

// String returns string representation of the [Token].
func (t Token) String() string {
	return string(t)
}
