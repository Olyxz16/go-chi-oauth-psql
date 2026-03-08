package services

import (
	"time"

	"github.com/Olyxz16/go-chi-oauth-psql/internal/auth/model"
	"github.com/Olyxz16/go-chi-oauth-psql/internal/config"
	"github.com/o1egl/paseto"
)

type TokenService struct {
	config *config.ServerConfig
	paseto *paseto.V2
}

func NewTokenService(cfg *config.ServerConfig) *TokenService {
	return &TokenService{
		config: cfg,
		paseto: paseto.NewV2(),
	}
}

func (s *TokenService) GenerateTokens(user *model.User) (string, string, error) {
	now := time.Now()
	exp := now.Add(24 * time.Hour)
	nbt := now

	jsonToken := paseto.JSONToken{
		Audience:   "cli-app",
		Issuer:     "go-chi-oauth-psql",
		Jti:        user.ID.String(),
		Subject:    user.Email,
		IssuedAt:   now,
		Expiration: exp,
		NotBefore:  nbt,
	}

	// Access Token
	accessToken, err := s.paseto.Encrypt([]byte(s.config.TokenSecret), jsonToken, "access_token")
	if err != nil {
		return "", "", err
	}

	// Refresh Token
	refreshExp := now.Add(7 * 24 * time.Hour)
	jsonToken.Expiration = refreshExp
	refreshToken, err := s.paseto.Encrypt([]byte(s.config.TokenSecret), jsonToken, "refresh_token")
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *TokenService) ValidateAccessToken(tokenStr string) (*paseto.JSONToken, error) {
	var token paseto.JSONToken
	var footer string
	err := s.paseto.Decrypt(tokenStr, []byte(s.config.TokenSecret), &token, &footer)
	if err != nil {
		return nil, err
	}

	if footer != "access_token" {
		return nil, paseto.ErrInvalidTokenAuth
	}

	if err := token.Validate(paseto.IssuedBy("go-chi-oauth-psql"), paseto.ForAudience("cli-app")); err != nil {
		return nil, err
	}

	return &token, nil
}

func (s *TokenService) ValidateRefreshToken(tokenStr string) (*paseto.JSONToken, error) {
	var token paseto.JSONToken
	var footer string
	err := s.paseto.Decrypt(tokenStr, []byte(s.config.TokenSecret), &token, &footer)
	if err != nil {
		return nil, err
	}

	if footer != "refresh_token" {
		return nil, paseto.ErrInvalidTokenAuth
	}

	if err := token.Validate(paseto.IssuedBy("go-chi-oauth-psql"), paseto.ForAudience("cli-app")); err != nil {
		return nil, err
	}

	return &token, nil
}
