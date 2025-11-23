package auth

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"p2m-lite/config"
	"p2m-lite/database"
	"p2m-lite/internal/utils"
)

type AuthService struct {
    cfg *config.Config
    db  *database.MockDB
}

func NewService(cfg *config.Config, db *database.MockDB) *AuthService {
    return &AuthService{
        cfg: cfg,
        db:  db,
    }
}

func (s *AuthService) InitiateChallenge(publicKey string) (randomString, base64Value string, err error) {
    log.Println("AUTH: Starting InitiateChallenge.")

    challengeBytes, err := utils.GenerateRandomBytes(32)
    if err != nil {
        return "", "", errors.New("failed to generate random challenge")
    }
    randomString = fmt.Sprintf("%x", challengeBytes)

    timestamp := time.Now().Unix()
    timestampStr := strconv.FormatInt(timestamp, 10)

    base64Value, err = utils.EncodeChallengeData(publicKey, randomString, timestampStr, s.cfg.AppSecret)
    if err != nil {
        return "", "", fmt.Errorf("failed to encode challenge data: %w", err)
    }

    log.Printf("AUTH: Challenge created. Challenge: %s, Timestamp: %s", randomString, timestampStr)

    if !s.db.PublicKeyExists(publicKey) {
        log.Println("AUTH: New public key detected. Challenge initiated.")
    } else {
        log.Println("AUTH: Known public key detected. Challenge initiated.")
    }

    log.Printf("AUTH: Challenge initiated. Challenge: %s, KeyData: %s...", randomString[:10], base64Value[:10])

    return randomString, base64Value, nil
}

func (s *AuthService) CompleteAuthAndIssueToken(signedChallengeB64, keyDataB64 string) (sessionToken string, err error) {
    log.Println("AUTH: Starting Verification.")

    decodedData, err := utils.DecodeAndExtractChallengeData(keyDataB64, s.cfg.AppSecret)
    if err != nil {
        return "", fmt.Errorf("decryption/decoding failed: %w", err)
    }

    publicKeyPEM := decodedData.PublicKey
    originalChallengeHex := decodedData.ChallengeString
    timestampStr := decodedData.TimestampStr

    log.Println("AUTH: Extracted Public Key PEM (first 5 lines):")
    log.Println(strings.Join(strings.Split(publicKeyPEM, "\n")[:5], "\n"))

    challengeTime, err := strconv.ParseInt(timestampStr, 10, 64)
    if err != nil {
        return "", errors.New("invalid timestamp format in key data")
    }

    expiryTime := time.Unix(challengeTime, 0).Add(time.Duration(s.cfg.SecretTTL) * time.Second)
    if time.Now().After(expiryTime) {
        log.Printf("AUTH: Challenge rejected. Expired at %s (Secret TTL: %d seconds).", expiryTime.Format(time.RFC3339), s.cfg.SecretTTL)
        return "", errors.New("authentication challenge expired")
    }
    log.Println("AUTH: Timestamp check PASSED.")

    block, _ := pem.Decode([]byte(publicKeyPEM))
    if block == nil {
        return "", errors.New("failed to decode public key PEM")
    }
    pub, err := x509.ParsePKIXPublicKey(block.Bytes)
    if err != nil {
        return "", fmt.Errorf("failed to parse public key: %w", err)
    }

    rsaPubKey, ok := pub.(*rsa.PublicKey)
    if !ok {
        return "", errors.New("public key is not an RSA public key")
    }
    fmt.Println("AUTH: Parsed RSA Public Key successfully.")

    signature, err := utils.Base64Decode(signedChallengeB64)
    if err != nil {
        return "", errors.New("invalid base64 signed challenge")
    }

    challengeBytes := []byte(originalChallengeHex)
    hashedChallenge := sha256.Sum256(challengeBytes)

    if err = rsa.VerifyPKCS1v15(rsaPubKey, crypto.SHA256, hashedChallenge[:], signature); err != nil {
        log.Printf("AUTH: Signature verification FAILED: %v", err)
        return "", errors.New("signature verification failed. The challenge was not signed by the private key")
    }
    log.Println("AUTH: Signature verification PASSED.")

    s.db.SavePublicKey(publicKeyPEM)

    sessionToken, err = utils.GenerateToken(publicKeyPEM, s.cfg.AppSecret, s.cfg.TokenTTL)
    if err != nil {
        return "", errors.New("failed to generate session token")
    }

    expiry := time.Now().Add(time.Duration(s.cfg.TokenTTL) * time.Second)
    s.db.StoreSession(publicKeyPEM, sessionToken, expiry)

    log.Printf("AUTH: Token issued. Token TTL: %d seconds.", s.cfg.TokenTTL)

    return sessionToken, nil
}
