// Package option defines a generic option function.
package option

// Option represents a function to configure a pointer to an object of given type.
type Option[T any] func(*T)
