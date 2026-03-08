package controller

import (
	"net/http"

	"github.com/Olyxz16/go-chi-oauth-psql/internal/api/middlewares"
	"github.com/Olyxz16/go-chi-oauth-psql/internal/auth/services"
	"github.com/go-chi/chi/v5"
	"github.com/markbates/goth/gothic"
)

func AuthController(userService *services.UserService, tokenService *services.TokenService, googleClientID string) *chi.Mux {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		gothic.BeginAuthHandler(w, r)
	})
	r.Get("/google/callback", HandleGoogleCallback(userService))
	r.Post("/google", HandleGoogleCLI(userService, tokenService, googleClientID))
	r.Post("/refresh", HandleRefresh(userService, tokenService))

	r.Group(func(r chi.Router) {
		r.Use(middlewares.AuthMiddleware(tokenService))
		r.Get("/me", HandleWhoAmI())
	})

	return r
}
