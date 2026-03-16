package storage

import "context"

// Object describes one immutable stored blob.
type Object struct {
	ID               string
	HouseholdID      string
	OriginalFilename string
	Checksum         string
	MimeType         string
	SizeBytes        int64
	Content          []byte
	Created          bool
}

// PutInput creates or reuses a stored object.
type PutInput struct {
	HouseholdID      string
	OriginalFilename string
	MimeType         string
	Checksum         string
	Content          []byte
}

// Store persists and retrieves immutable blob content.
type Store interface {
	Put(ctx context.Context, in PutInput) (Object, error)
	Get(ctx context.Context, householdID string, objectID string) (Object, error)
	Delete(ctx context.Context, householdID string, objectID string) error
}
