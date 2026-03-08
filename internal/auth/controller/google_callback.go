package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Olyxz16/go-chi-oauth-psql/internal/auth/model"
	"github.com/Olyxz16/go-chi-oauth-psql/internal/auth/services"
	"github.com/markbates/goth/gothic"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type AuthRequest struct {
	Code        string `json:"code"`
	RedirectURI string `json:"redirect_uri"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func HandleGoogleCallback(svc *services.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gu, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			http.Error(w, "Authentication failed", http.StatusForbidden)
			return
		}

		user, err := svc.CreateUser(r.Context(), gu.Email, model.Google)
		if err != nil {
			log.Printf("Error creating/reading user in DB: %v", err)
			http.Error(w, "Internal server error during user registration", http.StatusInternalServerError)
			return
		}

		s, err := gothic.Store.New(r, "session")
		if err != nil {
			log.Printf("Error getting new session: %v", err)
			http.Error(w, "Session error", http.StatusInternalServerError)
			return
		}

		s.Values["user"] = user
		err = s.Save(r, w)
		if err != nil {
			log.Printf("Error saving session: %v", err)
			http.Error(w, "Session save error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func HandleGoogleCLI(userService *services.UserService, tokenService *services.TokenService, googleClientID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AuthRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Configure OAuth2 with server credentials
		conf := &oauth2.Config{
			ClientID:     googleClientID,
			ClientSecret: os.Getenv("GOOGLE_SECRET_ACCESS_KEY"), // Must be set in API env
			Endpoint:     google.Endpoint,
			RedirectURL:  req.RedirectURI, // MUST match what the CLI used
		}

		// Exchange code for token
		token, err := conf.Exchange(context.Background(), req.Code)
		if err != nil {
			log.Printf("Token exchange failed: %v", err)
			http.Error(w, fmt.Sprintf("Token exchange failed: %v", err), http.StatusUnauthorized)
			return
		}

		// Use Access Token to fetch user info directly
		client := conf.Client(context.Background(), token)
		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
		if err != nil {
			http.Error(w, "Failed to get user info", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var userInfo struct {
			Email string `json:"email"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
			http.Error(w, "Failed to parse user info", http.StatusInternalServerError)
			return
		}

		user, err := userService.CreateUser(r.Context(), userInfo.Email, model.Google)
		if err != nil {
			log.Printf("Error creating/getting user: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		accessToken, refreshToken, err := tokenService.GenerateTokens(user)
		if err != nil {
			log.Printf("Error generating tokens: %v", err)
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
