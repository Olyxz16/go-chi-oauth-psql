package config

import (
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
)

func SetupGoth(cfg *GothConfig) {
	goth.UseProviders(
		google.New(
			cfg.GoogleAccessKeyId,
			cfg.GoogleSecretAccessKey,
			"http://localhost:8080/auth/google/callback",
			"email",
		),
	)
}
