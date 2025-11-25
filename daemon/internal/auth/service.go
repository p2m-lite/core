package auth

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Service struct {
	client *http.Client
}

func NewService() *Service {
	return &Service{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *Service) InitiateChallenge(url string, publicKeyPEM string) (string, string, error) {
	reqPayload := map[string]string{
		"public_key": publicKeyPEM,
	}
	b, err := json.Marshal(reqPayload)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal JSON payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("server returned error status: %s, body: %s", resp.Status, string(body))
	}

	var responseData struct {
		Challenge string `json:"challenge"`
		KeyData   string `json:"key_data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return "", "", fmt.Errorf("failed to decode response: %w", err)
	}

	return responseData.Challenge, responseData.KeyData, nil
}

func (s *Service) VerifyAuth(url string, privateKey *ecdsa.PrivateKey, challenge string, keyData string) (string, error) {
	challengeBytes := []byte(challenge)
	hashed := sha256.Sum256(challengeBytes)

	r, rs, err := ecdsa.Sign(rand.Reader, privateKey, hashed[:])
	if err != nil {
		return "", fmt.Errorf("failed to sign challenge: %w", err)
	}

	const keyBytes = 32
	rBytes := r.Bytes()
	sBytes := rs.Bytes()

	signature := make([]byte, keyBytes*2)
	copy(signature[keyBytes-len(rBytes):keyBytes], rBytes)
	copy(signature[keyBytes*2-len(sBytes):keyBytes*2], sBytes)

	signedChallengeB64 := base64.URLEncoding.EncodeToString(signature)

	reqPayload := map[string]string{
		"signed_challenge": signedChallengeB64,
		"key_data":         keyData,
	}
	b, err := json.Marshal(reqPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("server returned error status: %s, body: %s", resp.Status, string(body))
	}

	var responseData struct {
		SessionToken string `json:"session_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return responseData.SessionToken, nil
}
