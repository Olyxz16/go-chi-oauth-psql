package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-redis/redis_rate/v10"

	"github.com/Olyxz16/go-chi-oauth-psql/internal/auth/controller"
	"github.com/Olyxz16/go-chi-oauth-psql/internal/auth/services"
	"github.com/Olyxz16/go-chi-oauth-psql/internal/api/middlewares"
)

func RegisterRoutes(userService *services.UserService, tokenService *services.TokenService, googleClientID string, limiter *redis_rate.Limiter) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Group(func (r chi.Router) {
		r.Use(middleware.Logger)
		r.Use(middlewares.RateLimitMiddleware(limiter))
		r.Mount("/auth", controller.AuthController(userService, tokenService, googleClientID))
	})

	return r
}
