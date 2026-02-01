package sqldatabase

import (
	"fmt"

	"gorm.io/gorm"
)

// TransactionFunc defines a function to be executed inside a database transaction.
type TransactionFunc func(*Storage) error

// Transaction wraps [TransactionFunc] to execute inside a database transaction.
type Transaction struct {
	fn TransactionFunc
}

// Run executes [TransactionFunc] inside a database transaction.
// If error is returned by [TransactionFunc], the transaction will be rolled back.
func (t Transaction) Run(s *Storage) error {
	err := s.db.Transaction(func(trx *gorm.DB) error {
		stx := &Storage{
			db: trx,
		}

		return t.fn(stx)
	})
	if err != nil {
		return fmt.Errorf("cannot execute transaction: %w", err)
	}

	return nil
}
