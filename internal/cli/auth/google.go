package auth

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func PerformGoogleLogin(googleClientId string) (string, string, error) {

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", "", fmt.Errorf("failed to start local listener: %w", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	redirectURL := fmt.Sprintf("http://127.0.0.1:%d/", port)

	conf := &oauth2.Config{
		ClientID:     googleClientId,
		ClientSecret: "", // CLI clients typically don't use a secret, or it's not needed for the code flow here
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
		RedirectURL:  redirectURL,
	}

	// 3. Generate the Google Auth URL
	state := "state-token" // In production, use a random string for security
	authURL := conf.AuthCodeURL(state, oauth2.AccessTypeOffline)

	fmt.Printf("Opening browser for Google Login...\n")
	fmt.Printf("If it doesn't open, visit:\n%s\n", authURL)

	// 4. Set up the server to handle the callback
	codeChan := make(chan string)
	errChan := make(chan error)

	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				http.NotFound(w, r)
				return
			}

			// Validate state
			if r.URL.Query().Get("state") != state {
				http.Error(w, "Invalid state parameter", http.StatusBadRequest)
				errChan <- fmt.Errorf("invalid state parameter")
				return
			}

			// Get code
			code := r.URL.Query().Get("code")
			if code == "" {
				http.Error(w, "Code not found", http.StatusBadRequest)
				errChan <- fmt.Errorf("code not found")
				return
			}

			// Respond to user
			w.Write([]byte("Login successful! You can close this window."))
			codeChan <- code
		}),
	}

	// 5. Run the server in a goroutine
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// 6. Wait for the code or error
	var code string
	select {
	case code = <-codeChan:
	case err := <-errChan:
		return "", "", err
	case <-time.After(5 * time.Minute): // Timeout
		return "", "", fmt.Errorf("timeout waiting for login")
	}

	// Graceful shutdown of server
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	server.Shutdown(ctx)

	// Return the code AND the redirectURL so the API can use the exact same one for exchange
	return code, redirectURL, nil
}
