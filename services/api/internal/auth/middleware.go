package auth

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	apierrors "github.com/libpulse/platform/services/api/internal/utils/errors"
)

// This will be used as the key in Gin Context claims
const ContextKeyClaims = "supabaseClaims"

// Supabase JWT Claims
type SupabaseClaims struct {
	Email       string `json:"email"`
	Role        string `json:"role"`
	AppMetadata struct {
		Provider string `json:"provider"`
	} `json:"app_metadata"`
	jwt.RegisteredClaims
}

// ErrorResponse（aligned with OpenAPI ErrorResponse）
type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

// NewMiddleware will return a Gin middleware to verify Supabase JWT
func NewMiddleware(jwtSecret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check Authorization Header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			apiErr := apierrors.NewAPIError(apierrors.ErrUnauthorized)
			c.AbortWithStatusJSON(apiErr.StatusCode(), apiErr)
			return
		}

		tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))

		token, err := jwt.ParseWithClaims(tokenStr, &SupabaseClaims{}, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return jwtSecret, nil
		})

		// Check whether the token is valid
		if err != nil || !token.Valid {
			apiErr := apierrors.NewAPIError(apierrors.ErrInvalidToken)
			c.AbortWithStatusJSON(apiErr.StatusCode(), apiErr)
			return
		}

		// Check the claims structure.
		claims, ok := token.Claims.(*SupabaseClaims)
		if !ok {
			apiErr := apierrors.NewAPIError(apierrors.ErrInvalidToken)
			c.AbortWithStatusJSON(apiErr.StatusCode(), apiErr)
			return
		}

		// Set the claims in context, so handlers can get it.
		c.Set(ContextKeyClaims, claims)
		c.Next()
	}
}
