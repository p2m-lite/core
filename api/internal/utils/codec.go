package utils

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

const DataDelimiter = "|" 

type DecodedData struct {
    PublicKey      string
    ChallengeString string
    TimestampStr    string
}

func EncodeChallengeData(publicKey, challengeString, timestampStr, appSecret string) (base64Value string, err error) {
    combinedData := strings.Join([]string{publicKey, challengeString, timestampStr}, DataDelimiter)

    encryptedData, err := EncryptWithSecret(combinedData, appSecret)
    if err != nil {
        return "", errors.New("failed to encrypt combined data")
    }

    base64Value = Base64Encode(encryptedData)
    
    return base64Value, nil
}

func DecodeAndExtractChallengeData(keyDataB64, appSecret string) (*DecodedData, error) {
    encryptedData, err := Base64Decode(keyDataB64)
    if err != nil {
        return nil, errors.New("invalid base64 key data")
    }

    combinedData, err := DecryptWithSecret(encryptedData, appSecret)
    if err != nil {
        return nil, errors.New("failed to decrypt key data. Server secret mismatch or data corrupted")
    }
    
    parts := strings.Split(combinedData, DataDelimiter)
    if len(parts) != 3 {
        return nil, fmt.Errorf("decrypted data has incorrect format. Expected 3 parts, got %d", len(parts))
    }

    const (
        ansiReset  = "\033[0m"
        ansiCyan   = "\033[36m"
        ansiYellow = "\033[33m"
        ansiGreen  = "\033[32m"
    )
    fmt.Printf("%sUTILS:%s %sDecoded Data%s - %sPublicKey length:%s %s%d%s %sChallengeString:%s %s%q%s %sTimestampStr:%s %s%q%s\n",
        ansiCyan, ansiReset,
        ansiYellow, ansiReset,
        ansiCyan, ansiReset, ansiGreen, len(parts[0]), ansiReset,
        ansiCyan, ansiReset, ansiGreen, parts[1], ansiReset,
        ansiCyan, ansiReset, ansiGreen, parts[2], ansiReset,
    )
    
    return &DecodedData{
        PublicKey:      parts[0],
        ChallengeString: parts[1],
        TimestampStr:    parts[2],
    }, nil
}

func ParseWeb3PublicKey(publicKeyPEM string) (*ecdsa.PublicKey, error) {
    block, _ := pem.Decode([]byte(publicKeyPEM))
    if block == nil {
        return nil, errors.New("failed to decode public key PEM block")
    }

    switch block.Type {
    case "PUBLIC KEY":
        pub, err := x509.ParsePKIXPublicKey(block.Bytes)
        if err != nil {
            return nil, fmt.Errorf("failed to parse PKIX public key: %w", err)
        }
        ecdsaPubKey, ok := pub.(*ecdsa.PublicKey)
        if !ok {
            return nil, errors.New("PKIX public key is not an ECDSA key")
        }
        return ecdsaPubKey, nil
    case "SECP256K1 PUBLIC KEY":
        pubKey, err := crypto.UnmarshalPubkey(block.Bytes)
        if err != nil {
            return nil, fmt.Errorf("failed to unmarshal SECP256K1 public key bytes: %w", err)
        }
        return pubKey, nil
    default:
        return nil, fmt.Errorf("unsupported public key PEM block type: %s", block.Type)
    }
}