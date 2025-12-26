package handlers

import (
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/libpulse/platform/services/api/internal/auth"
	"github.com/libpulse/platform/services/api/internal/supabase"
	"github.com/libpulse/platform/services/api/internal/utils/crypto"
	"github.com/libpulse/platform/services/api/internal/utils/errors"
)

// CreateProjectRequest matches the OpenAPI schema
type CreateProjectRequest struct {
	Name string `json:"name" binding:"required,min=1,max=64"`
}

// CreateProjectResponse matches the OpenAPI schema
type CreateProjectResponse struct {
	ID string `json:"id"`
}

// CreateProjectHandler handles POST /api/v1/projects
func CreateProjectHandler(store ProjectStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1) ensure claims (injected by auth middleware)
		claimsAny, ok := c.Get(auth.ContextKeyClaims)
		if !ok {
			apiErr := errors.NewAPIError(errors.ErrUnauthorized)
			c.JSON(apiErr.StatusCode(), apiErr)
			return
		}

		// 2) Check the claims type
		claims, ok := claimsAny.(*auth.SupabaseClaims)
		if !ok || claims.Subject == "" {
			apiErr := errors.NewAPIError(errors.ErrUnauthorized)
			c.JSON(apiErr.StatusCode(), apiErr)
			return
		}

		// 3) Parse and validate request body
		var req CreateProjectRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			apiErr := errors.NewAPIError(errors.ErrBadRequest)
			c.JSON(apiErr.StatusCode(), apiErr)
			return
		}

		// 4) Create project
		project, err := store.CreateProject(c.Request.Context(), req.Name, claims.Subject)
		if err != nil {
			// Check if it's a duplicate project name error
			errMsg := strings.ToLower(err.Error())

			// Log the error for debugging
			log.Printf("CreateProject error: %s", err.Error())

			// Check for various unique constraint violation patterns
			if strings.Contains(errMsg, "duplicate") ||
			   strings.Contains(errMsg, "unique constraint") ||
			   strings.Contains(errMsg, "unique_violation") ||
			   strings.Contains(errMsg, "23505") || // PostgreSQL unique violation code
			   strings.Contains(errMsg, "projects_owner_name_unique") {
				apiErr := errors.NewAPIError(errors.ErrConflict)
				c.JSON(apiErr.StatusCode(), apiErr)
				return
			}

			// Other database errors
			apiErr := errors.NewAPIError(errors.ErrInternalError)
			c.JSON(apiErr.StatusCode(), apiErr)
			return
		}

		// 5) Defensive: store returned (nil, nil), which should never happen
		if project == nil {
			apiErr := errors.NewAPIError(errors.ErrInternalError)
			c.JSON(apiErr.StatusCode(), apiErr)
			return
		}

		// 6) Return response
		c.JSON(http.StatusCreated, CreateProjectResponse{
			ID: project.ID,
		})
	}
}

// CreateProjectKeyRequest matches the OpenAPI schema
type CreateProjectKeyRequest struct {
	Label             string   `json:"label" binding:"required,min=1,max=64"`
	RequireSignature  bool     `json:"require_signature"`
	Env               *string  `json:"env"`
	Scopes            []string `json:"scopes"`
}

// ProjectKeyResponse matches the OpenAPI ProjectKey schema
type ProjectKeyResponse struct {
	ID          string    `json:"id"`
	Label       string    `json:"label"`
	Env         *string   `json:"env"`
	Scopes      []string  `json:"scopes"`
	SecretLast4 string    `json:"secret_last4"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateProjectKeyResponse matches the OpenAPI schema
type CreateProjectKeyResponse struct {
	ProjectKeyPublic string              `json:"project_key_public"`
	ProjectSecret    *string             `json:"project_secret"`
	Key              ProjectKeyResponse  `json:"key"`
}

// Simple in-memory rate limiter for key creation
type rateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
}

type bucket struct {
	count      int
	resetAt    time.Time
}

var (
	// Burst rate limiter: max 3 keys per minute
	keyCreationLimiterMinute = &rateLimiter{
		buckets: make(map[string]*bucket),
	}

	// Daily rate limiter: max 5 keys per day
	keyCreationLimiterDay = &rateLimiter{
		buckets: make(map[string]*bucket),
	}
)

const (
	maxKeysPerMinute = 3           // Burst protection: max 3 per minute
	maxKeysPerDay    = 5           // Daily quota: max 5 per day
	rateLimitWindowMinute = time.Minute
	rateLimitWindowDay    = 24 * time.Hour
)

func (rl *rateLimiter) allow(userID string, maxCount int, window time.Duration) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, exists := rl.buckets[userID]

	if !exists || now.After(b.resetAt) {
		rl.buckets[userID] = &bucket{
			count:   1,
			resetAt: now.Add(window),
		}
		return true
	}

	if b.count >= maxCount {
		return false
	}

	b.count++
	return true
}

// CreateProjectKeyHandler handles POST /api/v1/projects/{id}/keys
func CreateProjectKeyHandler(projectStore ProjectStore, keyStore ProjectKeyStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1) Ensure authentication
		claimsAny, ok := c.Get(auth.ContextKeyClaims)
		if !ok {
			apiErr := errors.NewAPIError(errors.ErrUnauthorized)
			c.JSON(apiErr.StatusCode(), apiErr)
			return
		}

		claims, ok := claimsAny.(*auth.SupabaseClaims)
		if !ok || claims.Subject == "" {
			apiErr := errors.NewAPIError(errors.ErrUnauthorized)
			c.JSON(apiErr.StatusCode(), apiErr)
			return
		}

		// 2) Extract project ID from URL
		projectID := c.Param("id")
		if projectID == "" {
			apiErr := errors.NewAPIError(errors.ErrBadRequest)
			c.JSON(apiErr.StatusCode(), apiErr)
			return
		}

		// 3) Rate limiting check - must pass BOTH limits
		// Check minute limit (burst protection)
		if !keyCreationLimiterMinute.allow(claims.Subject, maxKeysPerMinute, rateLimitWindowMinute) {
			apiErr := errors.NewAPIError(errors.ErrTooManyRequests)
			c.JSON(apiErr.StatusCode(), apiErr)
			return
		}

		// Check daily limit (quota protection)
		if !keyCreationLimiterDay.allow(claims.Subject, maxKeysPerDay, rateLimitWindowDay) {
			apiErr := errors.NewAPIError(errors.ErrTooManyRequests)
			c.JSON(apiErr.StatusCode(), apiErr)
			return
		}

		// 4) Parse and validate request body
		var req CreateProjectKeyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			apiErr := errors.NewAPIError(errors.ErrBadRequest)
			c.JSON(apiErr.StatusCode(), apiErr)
			return
		}

		// 5) Get project to verify it exists and check ownership
		project, err := projectStore.GetProjectByID(c.Request.Context(), projectID)
		if err != nil {
			errMsg := strings.ToLower(err.Error())
			if strings.Contains(errMsg, "not found") {
				apiErr := errors.NewAPIError(errors.ErrNotFound)
				c.JSON(apiErr.StatusCode(), apiErr)
				return
			}
			// Other errors (e.g., invalid UUID format) are parameter errors
			log.Printf("GetProjectByID error: %s", err.Error())
			apiErr := errors.NewAPIError(errors.ErrBadRequest)
			c.JSON(apiErr.StatusCode(), apiErr)
			return
		}

		if project == nil {
			apiErr := errors.NewAPIError(errors.ErrNotFound)
			c.JSON(apiErr.StatusCode(), apiErr)
			return
		}

		// 6) Verify user is the project owner
		if project.OwnerUserID != claims.Subject {
			apiErr := errors.NewAPIError(errors.ErrForbidden)
			c.JSON(apiErr.StatusCode(), apiErr)
			return
		}

		// 7) Set default values
		env := "prod"
		if req.Env != nil && *req.Env != "" {
			env = *req.Env
		}

		scopes := req.Scopes
		if len(scopes) == 0 {
			scopes = []string{"ingest"}
		}

		// 8) Generate keys
		publicKey, err := crypto.GeneratePublicKey()
		if err != nil {
			log.Printf("Failed to generate public key: %s", err.Error())
			apiErr := errors.NewAPIError(errors.ErrInternalError)
			c.JSON(apiErr.StatusCode(), apiErr)
			return
		}

		secret, err := crypto.GenerateSecret()
		if err != nil {
			log.Printf("Failed to generate secret: %s", err.Error())
			apiErr := errors.NewAPIError(errors.ErrInternalError)
			c.JSON(apiErr.StatusCode(), apiErr)
			return
		}

		// 9) Hash secret and get last4
		secretHash := crypto.HashSecret(secret)
		secretLast4 := crypto.GetLast4(secret)

		// 10) Create project key in database
		keyParams := supabase.CreateProjectKeyParams{
			ProjectID:   projectID,
			Label:       req.Label,
			Env:         env,
			SignedOnly:  req.RequireSignature,
			PublicKey:   publicKey,
			SecretHash:  secretHash,
			SecretLast4: secretLast4,
			CreatedBy:   claims.Subject,
		}

		projectKey, err := keyStore.CreateProjectKey(c.Request.Context(), keyParams)
		if err != nil {
			log.Printf("CreateProjectKey error: %s", err.Error())
			apiErr := errors.NewAPIError(errors.ErrInternalError)
			c.JSON(apiErr.StatusCode(), apiErr)
			return
		}

		if projectKey == nil {
			apiErr := errors.NewAPIError(errors.ErrInternalError)
			c.JSON(apiErr.StatusCode(), apiErr)
			return
		}

		// 11) Build response with secret (shown only once)
		c.Header("Cache-Control", "no-store")

		response := CreateProjectKeyResponse{
			ProjectKeyPublic: publicKey,
			ProjectSecret:    &secret,
			Key: ProjectKeyResponse{
				ID:          projectKey.ID,
				Label:       projectKey.Label,
				Env:         &projectKey.Env,
				Scopes:      scopes,
				SecretLast4: secretLast4,
				CreatedAt:   projectKey.CreatedAt,
			},
		}

		c.JSON(http.StatusCreated, response)
	}
}
