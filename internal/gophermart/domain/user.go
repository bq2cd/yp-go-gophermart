package domain

const (
	// UserIDEmptyValue represents empty value for [UserID].
	UserIDEmptyValue UserID = ""
)

// UserID represents a unique user identifier within the system.
// Typically, this is a login.
type UserID string
