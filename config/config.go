package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	ListenAddress      string
	DatabaseURL        string
	GitHubClientID     string
	GitHubClientSecret string
	GitHubRedirectURL  string
	InitialAdminUsers  []string
}

func Load() (*Config, error) {
	listenAddress := os.Getenv("LISTEN_ADDR")
	if listenAddress == "" {
		listenAddress = ":8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}

	githubClientID := strings.TrimSpace(os.Getenv("GITHUB_CLIENT_ID"))
	if githubClientID == "" {
		return nil, errors.New("GITHUB_CLIENT_ID is required")
	}

	githubClientSecret := strings.TrimSpace(os.Getenv("GITHUB_CLIENT_SECRET"))
	if githubClientSecret == "" {
		return nil, errors.New("GITHUB_CLIENT_SECRET is required")
	}

	githubRedirectURL := strings.TrimSpace(os.Getenv("GITHUB_REDIRECT_URL"))
	if githubRedirectURL == "" {
		return nil, errors.New("GITHUB_REDIRECT_URL is required")
	}

	initialAdmins := make([]string, 0)
	for _, admin := range strings.Split(strings.TrimSpace(os.Getenv("ADMIN_USERS")), ",") {
		admin = strings.ToLower(strings.TrimSpace(admin))
		if admin != "" {
			initialAdmins = append(initialAdmins, admin)
		}
	}

	return &Config{
		ListenAddress:      listenAddress,
		DatabaseURL:        databaseURL,
		GitHubClientID:     githubClientID,
		GitHubClientSecret: githubClientSecret,
		GitHubRedirectURL:  githubRedirectURL,
		InitialAdminUsers:  initialAdmins,
	}, nil
}

func (c *Config) String() string {
	return fmt.Sprintf("listen=%s db=%s", c.ListenAddress, c.DatabaseURL)
}
