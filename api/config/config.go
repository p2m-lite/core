package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

const (
	DefaultTokenTTL  = 600
	DefaultSecretTTL = 60
)

type Config struct {
	AppSecret       string
	TokenTTL        int
	SecretTTL       int
	BlockchainURL   string
	ContractAddress string
	BrevoAPIKey     string
	PrivateKey      string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Fatalf("Error loading .env file: %v", err)
	}

	appConfig := &Config{
		AppSecret:       os.Getenv("APP_SECRET"),
		TokenTTL:        DefaultTokenTTL,
		SecretTTL:       DefaultSecretTTL,
		BlockchainURL:   os.Getenv("BLOCKCHAIN_URL"),
		ContractAddress: os.Getenv("CONTRACT_ADDRESS"),
		BrevoAPIKey:     os.Getenv("BREVO_API_KEY"),
		PrivateKey:      os.Getenv("PRIVATE_KEY"),
	}

	if appConfig.AppSecret == "" {
		log.Fatal("APP_SECRET is not set in environment or .env file. The server cannot run without it.")
	}

	if appConfig.BlockchainURL == "" {
		log.Println("Warning: BLOCKCHAIN_URL is not set. WebSocket functionality may fail.")
	}

	if appConfig.ContractAddress == "" {
		log.Println("Warning: CONTRACT_ADDRESS is not set. WebSocket functionality may fail.")
	}

	if ttlStr := os.Getenv("TOKEN_TTL"); ttlStr != "" {
		if ttl, err := strconv.Atoi(ttlStr); err == nil {
			appConfig.TokenTTL = ttl
		} else {
			log.Printf("Warning: Invalid TOKEN_TTL value '%s'. Using default: %d seconds.", ttlStr, DefaultTokenTTL)
		}
	}

	if ttlStr := os.Getenv("SECRET_TTL"); ttlStr != "" {
		if ttl, err := strconv.Atoi(ttlStr); err == nil {
			appConfig.SecretTTL = ttl
		} else {
			log.Printf("Warning: Invalid SECRET_TTL value '%s'. Using default: %d seconds.", ttlStr, DefaultSecretTTL)
		}
	}

	log.Printf("Config loaded: Token TTL=%d, Secret TTL=%d", appConfig.TokenTTL, appConfig.SecretTTL)

	return appConfig
}
