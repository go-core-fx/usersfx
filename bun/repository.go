package bun

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/go-core-fx/usersfx"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Repository struct {
	db *bun.DB
}

func NewRepository(db *bun.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) RegisterOrLogin(ctx context.Context, ident usersfx.Identity) (*usersfx.User, error) {
	// First, try to find an existing existingUser with this identity
	existingUser, err := r.Login(ctx, ident)
	if err == nil {
		return existingUser, nil
	}

	if !errors.Is(err, usersfx.ErrNotFound) {
		return nil, err
	}

	// Create a new user
	newUser := newUser(uuid.NewString())

	// Create the identity
	newIdentity := newIdentity(
		newUser.ID,
		string(ident.Provider),
		ident.ProviderID,
		ident.ProviderData,
	)

	// Begin transaction
	err = r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err = tx.NewInsert().
			Model(newUser).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		if _, err = tx.NewInsert().
			Model(newIdentity).
			Exec(ctx); err != nil {
			// Check if the error is a duplicate key error (unique constraint violation)
			if isDuplicateKeyError(err) {
				// Another concurrent request created this identity, signal to fetch existing user
				return usersfx.ErrIdentityExists
			}
			return fmt.Errorf("failed to create identity: %w", err)
		}

		return nil
	})

	// If the identity already exists (race condition), fetch and return the existing user
	if errors.Is(err, usersfx.ErrIdentityExists) {
		return r.Login(ctx, ident)
	}

	if err != nil {
		return nil, fmt.Errorf("transaction failed: %w", err)
	}

	return newUser.toDomain(), nil
}

func (r *Repository) Login(ctx context.Context, ident usersfx.Identity) (*usersfx.User, error) {
	// Find user by identity
	var existingUser user
	err := r.db.NewSelect().
		Model(&existingUser).
		Where("id IN (?)", r.db.NewSelect().
			Model((*identity)(nil)).
			Column("user_id").
			Where("provider = ?", ident.Provider).
			Where("provider_id = ?", ident.ProviderID),
		).
		Scan(ctx, &existingUser)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, usersfx.ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return existingUser.toDomain(), nil
}

func (r *Repository) GetUser(ctx context.Context, userID string) (*usersfx.User, error) {
	var existingUser user
	err := r.db.NewSelect().
		Model(&existingUser).
		Where("id = ?", userID).
		Scan(ctx, &existingUser)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, usersfx.ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return existingUser.toDomain(), nil
}

// isDuplicateKeyError reports whether err is a unique-constraint violation
// as reported by SQLite, PostgreSQL or MySQL.
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") ||
		strings.Contains(msg, "duplicate key value violates unique constraint") ||
		strings.Contains(msg, "Duplicate entry")
}
