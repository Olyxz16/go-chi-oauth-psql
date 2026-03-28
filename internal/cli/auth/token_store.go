package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type TokenStore struct {
	appDir string
}

func NewTokenStore(appDir string) *TokenStore {
	return &TokenStore{
		appDir: appDir,
	}
}

func (m *TokenStore) getTokenPath() string {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".config", m.appDir)
	_ = os.MkdirAll(path, 0755)
	return filepath.Join(path, "tokens.json")
}

func (m *TokenStore) SaveTokens(t Tokens) error {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.getTokenPath(), data, 0600)
}

func (m *TokenStore) LoadTokens() (Tokens, error) {
	var t Tokens
	data, err := os.ReadFile(m.getTokenPath())
	if err != nil {
		return t, err
	}
	err = json.Unmarshal(data, &t)
	return t, err
}
