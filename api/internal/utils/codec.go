package utils

import (
	"errors"
	"fmt"
	"strings"
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
