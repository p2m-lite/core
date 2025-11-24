package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"
)

func GenerateRandomBytes(n int) ([]byte, error) {
    b := make([]byte, n)
    _, err := rand.Read(b)
    if err != nil {
        return nil, err
    }
    return b, nil
}

func Base64Encode(data []byte) string {
    return base64.URLEncoding.EncodeToString(data)
}

func Base64Decode(s string) ([]byte, error) {
    return base64.URLEncoding.DecodeString(s)
}

func EncryptWithSecret(data, secret string) ([]byte, error) {
    dataBytes := []byte(data)
    secretBytes := []byte(secret)
    if len(secretBytes) == 0 {
        return nil, errors.New("secret cannot be empty")
    }

    encrypted := make([]byte, len(dataBytes))
    for i := 0; i < len(dataBytes); i++ {
        encrypted[i] = dataBytes[i] ^ secretBytes[i%len(secretBytes)]
    }
    return encrypted, nil
}

func DecryptWithSecret(encryptedData []byte, secret string) (string, error) {
    secretBytes := []byte(secret)
    if len(secretBytes) == 0 {
        return "", errors.New("secret cannot be empty")
    }

    decrypted := make([]byte, len(encryptedData))
    for i := 0; i < len(encryptedData); i++ {
        decrypted[i] = encryptedData[i] ^ secretBytes[i%len(secretBytes)]
    }
    return string(decrypted), nil
}

func GenerateToken(publicKeyPEM, secret string, ttlSeconds int) (string, error) {
    expiry := time.Now().Add(time.Duration(ttlSeconds) * time.Second).Unix()
    payload := fmt.Sprintf(`{"pk_hash":"%x","exp":%d}`, sha256.Sum256([]byte(publicKeyPEM)), expiry)

    encodedPayload := Base64Encode([]byte(payload))

    hasher := sha256.New()
    hasher.Write([]byte(encodedPayload + secret))
    signature := hasher.Sum(nil)

    encodedSignature := Base64Encode(signature)

    return fmt.Sprintf("%s.%s", encodedPayload, encodedSignature), nil
}