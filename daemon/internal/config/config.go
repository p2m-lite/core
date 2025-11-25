package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AuthURL         string
	VerifyURL       string
	ContractAddress string
}

func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080" // Default
	}

	// Simple check to remove trailing slash if present to avoid double slashes
	if len(baseURL) > 0 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}

	return &Config{
		ContractAddress: os.Getenv("CONTRACT_ADDRESS"),
		AuthURL:   baseURL + "/auth/initiate",
		VerifyURL: baseURL + "/auth/verify",
	}, nil
}
