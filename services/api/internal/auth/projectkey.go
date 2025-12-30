package auth

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/libpulse/platform/services/api/internal/utils/crypto"
	"github.com/libpulse/platform/services/api/internal/utils/errors"
)

// ProjectKeyInfo contains the information needed from a project key
type ProjectKeyInfo interface {
	GetProjectID() string
	IsDisabled() bool
	RequiresSignature() bool
	GetSecretEncrypted() string // Returns encrypted secret for signature verification
}

// KeyLookupFunc is a function that looks up a project key by public key
type KeyLookupFunc func(c *gin.Context, publicKey string) (ProjectKeyInfo, error)

// NewProjectKeyMiddleware creates middleware that authenticates requests using X-Pulse-Key header
func NewProjectKeyMiddleware(lookupKey KeyLookupFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1) Extract X-Pulse-Key header
		pulseKey := c.GetHeader("X-Pulse-Key")
		if pulseKey == "" {
			apiErr := &errors.APIError{
				Error:  "Missing X-Pulse-Key header",
				Code:   errors.ErrUnauthorized,
				Status: 401,
			}
			c.JSON(apiErr.StatusCode(), apiErr)
			c.Abort()
			return
		}

		// 2) Validate key format (should start with pk_live_ )
		if !strings.HasPrefix(pulseKey, "pk_live_") {
			apiErr := &errors.APIError{
				Error:  "Invalid X-Pulse-Key format",
				Code:   errors.ErrUnauthorized,
				Status: 401,
			}
			c.JSON(apiErr.StatusCode(), apiErr)
			c.Abort()
			return
		}

		// 3) Lookup project key in database
		keyInfo, err := lookupKey(c, pulseKey)
		if err != nil {
			log.Printf("Failed to lookup project key: %v", err)
			apiErr := &errors.APIError{
				Error:  "Invalid X-Pulse-Key",
				Code:   errors.ErrUnauthorized,
				Status: 401,
			}
			c.JSON(apiErr.StatusCode(), apiErr)
			c.Abort()
			return
		}

		// 4) Check if key is disabled
		if keyInfo.IsDisabled() {
			apiErr := &errors.APIError{
				Error:  "Project key is disabled",
				Code:   errors.ErrUnauthorized,
				Status: 401,
			}
			c.JSON(apiErr.StatusCode(), apiErr)
			c.Abort()
			return
		}

		// 5) Check signature if required
		if keyInfo.RequiresSignature() {
			// Decrypt the secret for signature verification
			secretEncrypted := keyInfo.GetSecretEncrypted()
			secretPlain, err := crypto.DecryptSecret(secretEncrypted)
			if err != nil {
				log.Printf("Failed to decrypt secret: %v", err)
				apiErr := &errors.APIError{
					Error:  "Internal server error",
					Code:   errors.ErrInternalError,
					Status: 500,
				}
				c.JSON(apiErr.StatusCode(), apiErr)
				c.Abort()
				return
			}

			// Verify signature
			if apiErr := VerifyIngestSignature(c, &keyInfo, secretPlain); apiErr != nil {
				c.JSON(apiErr.StatusCode(), apiErr)
				c.Abort()
				return
			}
		}

		// 6) Set project_id in context for handler
		c.Set("project_id", keyInfo.GetProjectID())

		c.Next()
	}
}

func VerifyIngestSignature(c *gin.Context, keyInfo *ProjectKeyInfo, secretPlain string) *errors.APIError {
	ts := c.GetHeader("X-Pulse-Timestamp")
	sig := c.GetHeader("X-Pulse-Signature")

	if ts == "" || sig == "" {
		return errors.NewAPIError(errors.ErrUnauthorized) // or a more specific code
	}

	// parse signature: "v1=<hex>"
	if !strings.HasPrefix(sig, "v1=") {
		return errors.NewAPIError(errors.ErrUnauthorized)
	}
	sigHex := strings.TrimPrefix(sig, "v1=")

	// parse timestamp and replay window
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return errors.NewAPIError(errors.ErrUnauthorized)
	}
	if time.Since(t) > 5*time.Minute || time.Until(t) > 5*time.Minute {
		return errors.NewAPIError(errors.ErrUnauthorized)
	}

	// read raw body bytes
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return errors.NewAPIError(errors.ErrInternalError)
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(raw)) // restore for downstream

	// sha256 hex of raw bytes (gzip bytes if gzipped)
	sum := sha256.Sum256(raw)
	bodyHex := hex.EncodeToString(sum[:])

	// canonical path: actual path, no query
	method := strings.ToUpper(c.Request.Method)
	path := c.Request.URL.Path

	stringToSign := strings.Join([]string{method, path, ts, bodyHex}, "\n")

	// expected signature
	mac := hmac.New(sha256.New, []byte(secretPlain))
	mac.Write([]byte(stringToSign))
	expected := mac.Sum(nil)

	got, err := hex.DecodeString(sigHex)
	if err != nil {
		return errors.NewAPIError(errors.ErrUnauthorized)
	}
	if !hmac.Equal(got, expected) {
		return errors.NewAPIError(errors.ErrUnauthorized)
	}

	return nil
}
