package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Olyxz16/go-chi-oauth-psql/internal/cli/auth"
)

type AuthTransport struct {
	Base       http.RoundTripper
	BaseURL    string
	TokenStore *auth.TokenStore
}

func (t *AuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clonedReq := req.Clone(req.Context())

	tokens, err := t.TokenStore.LoadTokens()
	if err != nil {
		return nil, err
	}

	clonedReq.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	resp, err := t.Base.RoundTrip(clonedReq)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusUnauthorized {
		return resp, nil
	}

	// ---------------------------------------------------------
	// 401 DETECTED - START REFRESH FLOW
	// ---------------------------------------------------------

	resp.Body.Close()

	newTokens, refreshErr := t.performRefresh()
	if refreshErr != nil {
		// If refresh fails (e.g. user banned, refresh expired),
		// return the ORIGINAL 401 response so the CLI knows to ask for login.
		return resp, nil
	}

	err = t.TokenStore.SaveTokens(*newTokens)

	retryReq := req.Clone(req.Context())
	retryReq.Header.Set("Authorization", "Bearer "+newTokens.AccessToken)

	if req.GetBody != nil {
		body, _ := req.GetBody()
		retryReq.Body = body
	}

	return t.Base.RoundTrip(retryReq)
}

func (t *AuthTransport) performRefresh() (*auth.Tokens, error) {
	tokens, err := t.TokenStore.LoadTokens()
	if err != nil {
		return nil, err
	}

	reqBody, _ := json.Marshal(map[string]string{
		"refresh_token": tokens.RefreshToken,
	})

	refreshReq, _ := http.NewRequest("POST", t.BaseURL+"/auth/refresh", bytes.NewBuffer(reqBody))
	refreshReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(refreshReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh failed with status: %d", resp.StatusCode)
	}

	var newTokens auth.Tokens
	if err := json.NewDecoder(resp.Body).Decode(&newTokens); err != nil {
		return nil, err
	}

	return &newTokens, nil
}
