package sqldatabase

import (
	"fmt"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase/models"
)

// AutoMigrate applies necessary schema changes to the database.
//
//nolint:exhaustruct
func (s *Storage) AutoMigrate() error {
	err := s.db.AutoMigrate(
		models.User{},
		models.Order{},
		models.Balance{},
		models.Accrual{},
		models.Withdrawal{},
	)
	if err != nil {
		return fmt.Errorf("cannot run migrations: %w", err)
	}

	return nil
}
