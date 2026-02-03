package sqldatabase

import "errors"

var (
	// ErrUnsupportedDatabaseDriver is returned by [NewStorage] when it encounters a database driver
	// that is not supported.
	ErrUnsupportedDatabaseDriver = errors.New("unsupported database driver")
)
