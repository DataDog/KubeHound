package store

import (
	"strconv"
	"sync/atomic"
)

// idCounter is a monotonically increasing counter for generating unique IDs within a run.
var idCounter atomic.Int64

// ObjectID returns a unique int64 ID for use as a primary key.
// Named ObjectID for backward compatibility with callers.
func ObjectID() int64 {
	return idCounter.Add(1)
}

// ResetIDCounter resets the ID counter (for testing).
func ResetIDCounter() {
	idCounter.Store(0)
}

// Hex returns the hexadecimal string representation of an int64 ID,
// providing a string key suitable for use in maps and caches.
func Hex(id int64) string {
	return strconv.FormatInt(id, 16)
}
