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

	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080" // Default
	}

	// Simple check to remove trailing slash if present to avoid double slashes
	if len(apiURL) > 0 && apiURL[len(apiURL)-1] == '/' {
		apiURL = apiURL[:len(apiURL)-1]
	}

	return &Config{
		ContractAddress: os.Getenv("CONTRACT_ADDRESS"),
		AuthURL:   apiURL + "/auth/initiate",
		VerifyURL: apiURL + "/auth/verify",
	}, nil
}
