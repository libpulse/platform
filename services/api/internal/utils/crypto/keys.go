package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

var secretPepper string

// Init initializes the crypto package with the secret pepper from environment
func Init(pepper string) {
	secretPepper = pepper
}

// GeneratePublicKey generates a public key in format: pk_live_<random>
func GeneratePublicKey() (string, error) {
	randomBytes := make([]byte, 24) // 24 bytes = 32 chars base64
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(randomBytes)
	return fmt.Sprintf("pk_live_%s", encoded), nil
}

// GenerateSecret generates a secret key in format: psk_live_<random>
func GenerateSecret() (string, error) {
	randomBytes := make([]byte, 32) // 32 bytes = 43 chars base64
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(randomBytes)
	return fmt.Sprintf("psk_live_%s", encoded), nil
}

// HashSecret hashes a secret using HMAC-SHA256 with pepper and returns hex-encoded hash
func HashSecret(secret string) string {
	h := hmac.New(sha256.New, []byte(secretPepper))
	h.Write([]byte(secret))
	return hex.EncodeToString(h.Sum(nil))
}

// GetLast4 extracts the last 4 characters from a secret for display
func GetLast4(secret string) string {
	if len(secret) < 4 {
		return secret
	}
	return secret[len(secret)-4:]
}
