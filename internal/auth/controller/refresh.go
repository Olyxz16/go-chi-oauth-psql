package controller

import (
	"encoding/json"
	"net/http"

	"github.com/Olyxz16/go-chi-oauth-psql/internal/auth/services"
)

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func HandleRefresh(userService *services.UserService, tokenService *services.TokenService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RefreshRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate Refresh Token
		token, err := tokenService.ValidateRefreshToken(req.RefreshToken)
		if err != nil {
			http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
			return
		}

		// Verify user still exists
		// Token Subject is Email
		user, err := userService.GetUserByMail(r.Context(), token.Subject)
		if err != nil {
			http.Error(w, "User not found or access revoked", http.StatusUnauthorized)
			return
		}

		// Generate new pair
		accessToken, refreshToken, err := tokenService.GenerateTokens(user)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AuthResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		})
	}
}
