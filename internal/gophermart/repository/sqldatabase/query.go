package sqldatabase

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Query allows to perform arbitrary queries against the database
// without exposing [gorm.DB] instance directly.
func Query[T any](storage *Storage, opts ...clause.Expression) gorm.Interface[T] {
	return gorm.G[T](storage.db, opts...)
}
