package inmemory

import "github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"

// Order adds user ID information to [domain.Order].
// This is intended to be used for internal storage.
type Order struct {
	domain.Order

	UserID        domain.UserID
	AccrualPoints float64
}
