package main

import (
	"log"
	"time"

	"github.com/p2m-lite/core/daemon/internal/auth"
	"github.com/p2m-lite/core/daemon/internal/config"
	"github.com/p2m-lite/core/daemon/internal/key"
	"github.com/p2m-lite/core/daemon/internal/web3"
)

func main() {
	log.Println("Starting P2M Lite Daemon...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Printf("Configuration loaded. Auth URL: %s", cfg.AuthURL)

	keyManager, err := key.NewManager()
	if err != nil {
		log.Fatalf("Failed to initialize key manager: %v", err)
	}

	log.Println("Ensuring keys exist...")
	if err := keyManager.EnsureKeys(); err != nil {
		log.Fatalf("Failed to ensure keys: %v", err)
	}
	log.Println("Keys ensured.")

	authService := auth.NewService()

	if err := performAuth(cfg, keyManager, authService); err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	privKey, kErr := keyManager.GetPrivateKey()
	if kErr != nil {
		log.Fatalf("Failed to get private key: %v", kErr)
	}

	log.Println("Daemon is looping in background...")
	for {
		time.Sleep(10 * time.Second)
		// Generate a random ph value and turbidity for demo purposes
		ph := 7 + time.Now().Second()%3       // Example: 7, 8, or 9
		turbidity := 300 + time.Now().Second()%100 // Example: 300-399
		err := web3.SendLog(cfg.ContractAddress, privKey, ph, turbidity)
		if err != nil {
			log.Printf("Failed to send log to blockchain: %v", err)
			continue
		}
		
		log.Println("Log sent to blockchain.")
	}
}

func performAuth(cfg *config.Config, km *key.Manager, as *auth.Service) error {
	// Step 2: Initiate Challenge
	log.Println("Initiating Auth Challenge...")
	pubKeyPEM, err := km.GetPublicKeyPEM()
	if err != nil {
		return err
	}

	challenge, keyData, err := as.InitiateChallenge(cfg.AuthURL, pubKeyPEM)
	if err != nil {
		return err
	}
	log.Println("Challenge received.")

	// Step 3: Verify Auth
	log.Println("Verifying Auth...")
	privKey, err := km.GetPrivateKey()
	if err != nil {
		return err
	}

	token, err := as.VerifyAuth(cfg.VerifyURL, privKey, challenge, keyData)
	if err != nil {
		return err
	}

	log.Printf("Authentication successful! Session Token: %s...", token[:10]) // Log first 10 chars
	return nil
}
