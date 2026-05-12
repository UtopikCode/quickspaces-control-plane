package config

import (
	"errors"
	"fmt"
	"os"
)

type Config struct {
	ListenAddress     string
	DatabaseURL       string
	ExecutionProvider string
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

	executionProvider := os.Getenv("EXECUTION_PROVIDER")
	if executionProvider == "" {
		return nil, errors.New("EXECUTION_PROVIDER is required")
	}

	return &Config{
		ListenAddress:     listenAddress,
		DatabaseURL:       databaseURL,
		ExecutionProvider: executionProvider,
	}, nil
}

func (c *Config) String() string {
	return fmt.Sprintf("listen=%s provider=%s db=%s", c.ListenAddress, c.ExecutionProvider, c.DatabaseURL)
}
