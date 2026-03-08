package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Olyxz16/go-chi-oauth-psql/internal/cli/auth"
	"github.com/Olyxz16/go-chi-oauth-psql/internal/cli/client"
	"github.com/spf13/cobra"
)

func NewWhoAmICommand() *cobra.Command {
	var apiURL string

	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "Display the current logged in user",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Setup authenticated client
		
tokenStore := auth.NewTokenStore("go-chi-oauth-psql-cli")
			
			// Check if tokens exist first
			_, err := tokenStore.LoadTokens()
			if err != nil {
				return fmt.Errorf("not logged in. Run 'login' command first")
			}

		
transport := &client.AuthTransport{
				Base:       http.DefaultTransport,
				BaseURL:    apiURL,
				TokenStore: tokenStore,
			}
			
			httpClient := &http.Client{
				Transport: transport,
				Timeout:   10 * time.Second,
			}

			// Make request
			resp, err := httpClient.Get(apiURL + "/auth/me")
			if err != nil {
				return fmt.Errorf("failed to contact API: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusUnauthorized {
				return fmt.Errorf("session expired or invalid. Please login again")
			}
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("API returned error: %s", resp.Status)
			}

			var user struct {
				Email string `json:"email"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
				return fmt.Errorf("failed to decode response: %w", err)
			}

			fmt.Printf("Logged in as: %s\n", user.Email)
			return nil
		},
	}

	cmd.Flags().StringVar(&apiURL, "api-url", "http://localhost:8080", "API Base URL")

	return cmd
}
