package main

import (
	"fmt"
	"net/http"

	"github.com/Olyxz16/go-chi-oauth-psql/internal/api"
	"github.com/Olyxz16/go-chi-oauth-psql/internal/auth/repositories"
	"github.com/Olyxz16/go-chi-oauth-psql/internal/auth/services"
	"github.com/Olyxz16/go-chi-oauth-psql/internal/config"
)

func main() {
	gothConf := config.NewGothConfig()
	config.SetupGoth(gothConf)

	cfg := config.NewServerConfig()

	pgCfg := config.NewPostgresConfig()
	pool, err := config.NewPostgresPool(pgCfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to postgres: %v", err))
	}
	defer pool.Close()

	userRepo := repositories.NewUserRepository(pool)
	userService := services.NewUserService(userRepo)
	tokenService := services.NewTokenService(cfg)

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler: api.RegisterRoutes(userService, tokenService, gothConf.GoogleAccessKeyId),
	}

	server.ListenAndServe()
}

