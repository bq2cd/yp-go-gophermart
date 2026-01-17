package domain

import (
	"slices"
	"time"
)

// WithTimestamp exposes a method to get an object's timestamp, e.g.
// when the object was created or processed.
type WithTimestamp interface {
	Timestamp() time.Time
}

// SortByTimestampFromNewestToOldest performs in-place sorting of given items
// in reverse chronological order (from the newest to the oldest).
func SortByTimestampFromNewestToOldest[T WithTimestamp](items []T) {
	slices.SortStableFunc(items, func(a, b T) int {
		return b.Timestamp().Compare(a.Timestamp())
	})
}

// SortByTimestampFromOldestToNewest performs in-place sorting of given items
// in forward chronological order (from the oldest to the newest).
func SortByTimestampFromOldestToNewest[T WithTimestamp](items []T) {
	slices.SortStableFunc(items, func(a, b T) int {
		return a.Timestamp().Compare(b.Timestamp())
	})
}
