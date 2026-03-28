package main

import (
	"fmt"
	"net/http"

	"github.com/Olyxz16/go-chi-oauth-psql/internal/api"
	"github.com/Olyxz16/go-chi-oauth-psql/internal/auth/repositories"
	"github.com/Olyxz16/go-chi-oauth-psql/internal/auth/services"
	"github.com/Olyxz16/go-chi-oauth-psql/internal/config"
	"go.uber.org/zap"
)

func main() {
	gothConf := config.NewGothConfig()
	config.SetupGoth(gothConf)

	cfg := config.NewServerConfig()

	logger := config.DefaultLogger()

	pgCfg := config.NewPostgresConfig()
	pool, err := config.NewPostgresPool(pgCfg)
	if err != nil {
		logger.Fatal("Failed to connect to postgres", zap.Error(err))
	}
	defer pool.Close()

	userRepo := repositories.NewUserRepository(pool)
	userService := services.NewUserService(userRepo)
	tokenService := services.NewTokenService(cfg)

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler: api.RegisterRoutes(userService, tokenService, gothConf.GoogleAccessKeyId),
	}

	if err = server.ListenAndServe() ; err != nil {
		logger.Fatal("Server failed. ", zap.Error(err))
	}

}

