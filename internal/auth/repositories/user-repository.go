package repositories

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Olyxz16/go-chi-oauth-psql/internal/auth/model"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// CreateUser inserts a new user or does nothing if the ID already exists (idempotent upsert).
func (r *UserRepository) CreateUser(ctx context.Context, u *model.User) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO users (id, email, provider)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO NOTHING
	`, u.ID, u.Email, string(u.Provider))
	return err
}

// GetUser fetches a user by their UUID string.
func (r *UserRepository) GetUser(ctx context.Context, id string) (*model.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, email, provider
		FROM users
		WHERE id = $1
	`, id)

	var u model.User
	var provider string
	if err := row.Scan(&u.ID, &u.Email, &provider); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrUserNotFound
		}
		return nil, err
	}
	u.Provider = model.Provider(provider)
	return &u, nil
}

// DeleteUser removes a user by their UUID string.
func (r *UserRepository) DeleteUser(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM users WHERE id = $1
	`, id)
	return err
}
