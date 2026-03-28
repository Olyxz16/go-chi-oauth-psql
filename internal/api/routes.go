package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Olyxz16/go-chi-oauth-psql/internal/auth/controller"
	"github.com/Olyxz16/go-chi-oauth-psql/internal/auth/services"
)

func RegisterRoutes(userService *services.UserService, tokenService *services.TokenService, googleClientID string) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Mount("/auth", controller.AuthController(userService, tokenService, googleClientID))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return r
}
