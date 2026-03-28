package repositories

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Olyxz16/go-chi-oauth-psql/internal/auth/model"
	"github.com/Olyxz16/go-chi-oauth-psql/internal/db"
	"github.com/google/uuid"
)

type UserRepository struct {
	queries *db.Queries
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{queries: db.New(pool)}
}

// CreateUser inserts a new user or does nothing if the ID already exists (idempotent upsert).
func (r *UserRepository) CreateUser(ctx context.Context, u *model.User) error {
	var id pgtype.UUID
	copy(id.Bytes[:], u.ID[:])
	id.Valid = true

	err := r.queries.CreateUser(ctx, db.CreateUserParams{
		ID:       id,
		Email:    u.Email,
		Provider: string(u.Provider),
	})
	return err
}

// GetUser fetches a user by their UUID string.
func (r *UserRepository) GetUser(ctx context.Context, idStr string) (*model.User, error) {
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}

	var id pgtype.UUID
	copy(id.Bytes[:], parsedID[:])
	id.Valid = true

	row, err := r.queries.GetUser(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrUserNotFound
		}
		return nil, err
	}

	var uID uuid.UUID
	copy(uID[:], row.ID.Bytes[:])

	return &model.User{
		ID:       uID,
		Email:    row.Email,
		Provider: model.Provider(row.Provider),
	}, nil
}

// DeleteUser removes a user by their UUID string.
func (r *UserRepository) DeleteUser(ctx context.Context, idStr string) error {
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		return err
	}

	var id pgtype.UUID
	copy(id.Bytes[:], parsedID[:])
	id.Valid = true

	return r.queries.DeleteUser(ctx, id)
}
