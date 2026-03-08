package controller

import (
	"encoding/json"
	"net/http"

	"github.com/Olyxz16/go-chi-oauth-psql/internal/api/middlewares"
)

type UserResponse struct {
	Email string `json:"email"`
}

func HandleWhoAmI() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email, ok := r.Context().Value(middlewares.UserEmailKey).(string)
		if !ok {
			http.Error(w, "User email not found in context", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(UserResponse{
			Email: email,
		})
	}
}
