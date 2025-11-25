package key

import (
	"encoding/pem"

	"github.com/ethereum/go-ethereum/crypto"
)

// GenerateECCKeyPairWeb3 generates a new ECC secp256k1 key pair compatible with Web3.
// It returns the private and public keys in PEM format.
func GenerateECCKeyPairWeb3() ([]byte, []byte, error) {
	// Use go-ethereum's direct key generation, which defaults to secp256k1
	priv, err := crypto.GenerateKey()
	if err != nil {
		return nil, nil, err
	}

	// 1. Private Key Serialization (Raw 32-byte format, bypasses x509)
	privBytes := crypto.FromECDSA(priv)
	privPem := pem.EncodeToMemory(&pem.Block{Type: "SECP256K1 PRIVATE KEY", Bytes: privBytes})

	// 2. Public Key Serialization (Raw 65-byte uncompressed format, bypasses x509)
	pubBytes := crypto.FromECDSAPub(&priv.PublicKey)
	pubPem := pem.EncodeToMemory(&pem.Block{Type: "SECP256K1 PUBLIC KEY", Bytes: pubBytes})

	return privPem, pubPem, nil
}
