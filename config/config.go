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
	DatabaseName       string
	GitHubClientID     string
	GitHubClientSecret string
	GitHubRedirectURL  string
	InitialAccessRules []string
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

	databaseName := strings.TrimSpace(os.Getenv("DATABASE_NAME"))
	if databaseName == "" {
		databaseName = "quickspaces"
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

	initialRules := make([]string, 0)
	for _, rule := range strings.Split(strings.TrimSpace(os.Getenv("ADMIN_USERS")), ",") {
		rule = strings.ToLower(strings.TrimSpace(rule))
		if rule != "" {
			initialRules = append(initialRules, rule)
		}
	}

	return &Config{
		ListenAddress:      listenAddress,
		DatabaseURL:        databaseURL,
		DatabaseName:       databaseName,
		GitHubClientID:     githubClientID,
		GitHubClientSecret: githubClientSecret,
		GitHubRedirectURL:  githubRedirectURL,
		InitialAccessRules: initialRules,
	}, nil
}

func LoadDatabaseConfig() (*Config, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}

	databaseName := strings.TrimSpace(os.Getenv("DATABASE_NAME"))
	if databaseName == "" {
		databaseName = "quickspaces"
	}

	return &Config{
		DatabaseURL:  databaseURL,
		DatabaseName: databaseName,
	}, nil
}

func (c *Config) String() string {
	return fmt.Sprintf("listen=%s db=%s name=%s", c.ListenAddress, c.DatabaseURL, c.DatabaseName)
}
