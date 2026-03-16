package storage

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql/sqlgraph"

	"github.com/AndreasRoither/NomNomVault/backend/internal/ent"
	entgen "github.com/AndreasRoither/NomNomVault/backend/internal/ent/generated"
	"github.com/AndreasRoither/NomNomVault/backend/internal/ent/generated/storedobject"
)

// PostgresStore persists immutable blobs in PostgreSQL through Ent.
type PostgresStore struct {
	db *ent.Client
}

// NewPostgresStore creates a PostgreSQL-backed blob store.
func NewPostgresStore(db *ent.Client) *PostgresStore {
	return &PostgresStore{db: db}
}

// Put creates or reuses a stored object by checksum within a household.
func (s *PostgresStore) Put(ctx context.Context, in PutInput) (Object, error) {
	existing, err := s.db.StoredObject.Query().
		Where(
			storedobject.HouseholdIDEQ(in.HouseholdID),
			storedobject.ChecksumEQ(in.Checksum),
		).
		Only(ctx)
	if err == nil {
		return Object{
			ID:               existing.ID,
			HouseholdID:      existing.HouseholdID,
			OriginalFilename: existing.OriginalFilename,
			Checksum:         existing.Checksum,
			MimeType:         existing.MimeType,
			SizeBytes:        existing.SizeBytes,
			Content:          existing.Content,
			Created:          false,
		}, nil
	}
	if !entgen.IsNotFound(err) {
		return Object{}, fmt.Errorf("query stored object: %w", err)
	}

	objectEntity, err := s.db.StoredObject.Create().
		SetHouseholdID(in.HouseholdID).
		SetOriginalFilename(in.OriginalFilename).
		SetMimeType(in.MimeType).
		SetSizeBytes(int64(len(in.Content))).
		SetChecksum(in.Checksum).
		SetContent(in.Content).
		Save(ctx)
	if err != nil {
		if sqlgraph.IsConstraintError(err) {
			existing, queryErr := s.db.StoredObject.Query().
				Where(
					storedobject.HouseholdIDEQ(in.HouseholdID),
					storedobject.ChecksumEQ(in.Checksum),
				).
				Only(ctx)
			if queryErr == nil {
				return Object{
					ID:               existing.ID,
					HouseholdID:      existing.HouseholdID,
					OriginalFilename: existing.OriginalFilename,
					Checksum:         existing.Checksum,
					MimeType:         existing.MimeType,
					SizeBytes:        existing.SizeBytes,
					Content:          existing.Content,
					Created:          false,
				}, nil
			}
			if !entgen.IsNotFound(queryErr) {
				return Object{}, fmt.Errorf("query stored object after constraint collision: %w", queryErr)
			}
		}

		return Object{}, fmt.Errorf("save stored object: %w", err)
	}

	return Object{
		ID:               objectEntity.ID,
		HouseholdID:      objectEntity.HouseholdID,
		OriginalFilename: objectEntity.OriginalFilename,
		Checksum:         objectEntity.Checksum,
		MimeType:         objectEntity.MimeType,
		SizeBytes:        objectEntity.SizeBytes,
		Content:          objectEntity.Content,
		Created:          true,
	}, nil
}

// Get retrieves a stored object by household and ID.
func (s *PostgresStore) Get(ctx context.Context, householdID string, objectID string) (Object, error) {
	objectEntity, err := s.db.StoredObject.Query().
		Where(
			storedobject.IDEQ(objectID),
			storedobject.HouseholdIDEQ(householdID),
		).
		Only(ctx)
	if err != nil {
		return Object{}, fmt.Errorf("query stored object: %w", err)
	}

	return Object{
		ID:               objectEntity.ID,
		HouseholdID:      objectEntity.HouseholdID,
		OriginalFilename: objectEntity.OriginalFilename,
		Checksum:         objectEntity.Checksum,
		MimeType:         objectEntity.MimeType,
		SizeBytes:        objectEntity.SizeBytes,
		Content:          objectEntity.Content,
	}, nil
}

// Delete removes a stored object scoped to one household.
func (s *PostgresStore) Delete(ctx context.Context, householdID string, objectID string) error {
	affected, err := s.db.StoredObject.Delete().
		Where(
			storedobject.IDEQ(objectID),
			storedobject.HouseholdIDEQ(householdID),
		).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete stored object: %w", err)
	}
	if affected == 0 {
		return nil
	}
	return nil
}
