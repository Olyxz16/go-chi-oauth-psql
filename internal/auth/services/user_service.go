package services

import (
	"context"

	"github.com/Olyxz16/go-chi-oauth-psql/internal/auth/model"
	"github.com/Olyxz16/go-chi-oauth-psql/internal/auth/repositories"
	"github.com/google/uuid"
)

type UserService struct {
	repo *repositories.UserRepository
}

func NewUserService(repo *repositories.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) CreateUser(ctx context.Context, email string, provider model.Provider) (*model.User, error) {
	// Generate deterministic UUID from email
	id := uuid.NewMD5(uuid.NameSpaceURL, []byte(email))
	user := &model.User{
		ID:       id,
		Email:    email,
		Provider: provider,
	}

	err := s.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetUserById(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return s.repo.GetUser(ctx, id.String())
}

func (s *UserService) GetUserByMail(ctx context.Context, email string) (*model.User, error) {
	id := uuid.NewMD5(uuid.NameSpaceURL, []byte(email))
	return s.repo.GetUser(ctx, id.String())
}

func (s *UserService) DeleteUserById(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteUser(ctx, id.String())
}

func (s *UserService) DeleteUserByMail(ctx context.Context, email string) error {
	id := uuid.NewMD5(uuid.NameSpaceURL, []byte(email))
	return s.repo.DeleteUser(ctx, id.String())
}
