package key

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/crypto"
)

const (
	KeysDirName    = "keys"
	PrivateKeyFile = "private.pem"
	PublicKeyFile  = "public.pem"
	AppDataDirName = "p2m-lite"
)

type Manager struct {
	keysDir string
}

func NewManager() (*Manager, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user config directory: %w", err)
	}

	appDataDir := filepath.Join(configDir, AppDataDirName)
	keysDir := filepath.Join(appDataDir, KeysDirName)

	return &Manager{
		keysDir: keysDir,
	}, nil
}

func (m *Manager) EnsureKeys() error {
	if err := os.MkdirAll(m.keysDir, 0700); err != nil {
		return fmt.Errorf("failed to create keys directory: %w", err)
	}

	privPath := filepath.Join(m.keysDir, PrivateKeyFile)
	pubPath := filepath.Join(m.keysDir, PublicKeyFile)

	if m.fileExists(privPath) && m.fileExists(pubPath) {
		return nil // Keys already exist
	}

	// Generate new keys
	privPem, pubPem, err := GenerateECCKeyPairWeb3()
	if err != nil {
		return fmt.Errorf("failed to generate keys: %w", err)
	}

	if err := os.WriteFile(privPath, privPem, 0600); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}
	if err := os.WriteFile(pubPath, pubPem, 0644); err != nil {
		return fmt.Errorf("failed to write public key: %w", err)
	}

	return nil
}

func (m *Manager) GetPrivateKey() (*ecdsa.PrivateKey, error) {
	privPath := filepath.Join(m.keysDir, PrivateKeyFile)
	privBytes, err := os.ReadFile(privPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	block, _ := pem.Decode(privBytes)
	if block == nil {
		return nil, errors.New("failed to decode PEM block from private key file")
	}

	switch block.Type {
	case "PRIVATE KEY":
		k, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKCS#8 private key: %w", err)
		}
		pk, ok := k.(*ecdsa.PrivateKey)
		if !ok {
			return nil, errors.New("parsed PKCS#8 key is not ECDSA")
		}
		return pk, nil
	case "EC PRIVATE KEY":
		pk, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse EC private key: %w", err)
		}
		return pk, nil
	case "SECP256K1 PRIVATE KEY":
		privKey, err := crypto.ToECDSA(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to convert raw bytes to ECDSA private key: %w", err)
		}
		return privKey, nil
	default:
		return nil, fmt.Errorf("unsupported PEM block type: %s", block.Type)
	}
}

func (m *Manager) GetPublicKeyPEM() (string, error) {
	pubPath := filepath.Join(m.keysDir, PublicKeyFile)
	pubBytes, err := os.ReadFile(pubPath)
	if err != nil {
		return "", fmt.Errorf("failed to read public key file: %w", err)
	}
	return string(pubBytes), nil
}

func (m *Manager) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
