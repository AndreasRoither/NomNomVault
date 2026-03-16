package idgen

import "github.com/google/uuid"

// Generator creates externally visible IDs.
type Generator interface {
	New() string
}

// UUIDGenerator creates UUID-based string IDs.
type UUIDGenerator struct{}

// New returns a new UUID string.
func (UUIDGenerator) New() string {
	return uuid.NewString()
}
