package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/Olyxz16/go-chi-oauth-psql/internal/cli/auth"
	"github.com/spf13/cobra"
)

func NewLoginCommand() *cobra.Command {
	var clientID string
	var apiURL string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login with Google",
		RunE: func(cmd *cobra.Command, args []string) error {
			if clientID == "" {
				return fmt.Errorf("client-id is required (or GOOGLE_ACCESS_KEY_ID env var)")
			}

			fmt.Println("Starting Google Login...")
			code, redirectURI, err := auth.PerformGoogleLogin(clientID)
			if err != nil {
				return fmt.Errorf("failed to login with Google: %w", err)
			}
			fmt.Println("Got Authorization Code from Google. Exchanging for API tokens...")

			// Exchange token
			reqBody, _ := json.Marshal(map[string]string{
				"code":         code,
				"redirect_uri": redirectURI,
			})

			resp, err := http.Post(apiURL+"/auth/google", "application/json", bytes.NewBuffer(reqBody))
			if err != nil {
				return fmt.Errorf("failed to contact API: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("API login failed with status: %d", resp.StatusCode)
			}

			var tokens auth.Tokens
			if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
				return fmt.Errorf("failed to decode API response: %w", err)
			}

			// Save tokens
			store := auth.NewTokenStore("go-chi-oauth-psql-cli")
			if err := store.SaveTokens(tokens); err != nil {
				return fmt.Errorf("failed to save tokens: %w", err)
			}

			fmt.Println("Login successful! Tokens saved.")
			return nil
		},
	}

	cmd.Flags().StringVar(&clientID, "client-id", os.Getenv("GOOGLE_ACCESS_KEY_ID"), "Google Access key ID")
	cmd.Flags().StringVar(&apiURL, "api-url", "http://localhost:8080", "API Base URL")

	return cmd
}
