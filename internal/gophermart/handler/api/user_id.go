package api

const (
	userIDContextKeyName userIDContextKeyType = "api_user_id"
)

const (
	// UserIDEmptyValue represents empty value for [UserID].
	UserIDEmptyValue UserID = ""
)

type userIDContextKeyType string

// UserID contains the ID of a user and is passed via request context.
type UserID string

// IsEmpty returns [true] if current user ID is not an empty value.
func (u UserID) IsEmpty() bool {
	return u == UserIDEmptyValue
}
