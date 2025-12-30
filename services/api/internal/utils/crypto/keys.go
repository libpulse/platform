package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
)

var secretPepper string
var encryptionKey []byte

// Init initializes the crypto package with the secret pepper and master key from environment
func Init(pepper, masterKey string) {
	secretPepper = pepper
	// Use master key directly for AES-256-GCM encryption (must be 32 bytes)
	h := sha256.Sum256([]byte(masterKey))
	encryptionKey = h[:]
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

// EncryptSecret encrypts a secret using AES-256-GCM and returns base64-encoded ciphertext
func EncryptSecret(secret string) (string, error) {
	if len(encryptionKey) == 0 {
		return "", errors.New("encryption key not initialized")
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(secret), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptSecret decrypts a base64-encoded ciphertext and returns the plain secret
func DecryptSecret(encrypted string) (string, error) {
	if len(encryptionKey) == 0 {
		return "", errors.New("encryption key not initialized")
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// VerifySecret checks if a plain secret matches the stored hash (for backward compatibility)
func VerifySecret(plain, hashed string) bool {
	computed := HashSecret(plain)
	return hmac.Equal([]byte(computed), []byte(hashed))
}
