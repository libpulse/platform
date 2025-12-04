package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Error: "Missing or invalid Authorization header",
				Code:  "Unauthorized",
			})
			return
		}

		tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))

		token, err := jwt.ParseWithClaims(tokenStr, &SupabaseClaims{}, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("Unexpected signing method")
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Error: "Invalid token",
				Code:  "Unauthorized",
			})
			return
		}

		claims, ok := token.Claims.(*SupabaseClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Error: "invalid token claims",
				Code:  "unauthorized",
			})
			return
		}

		// Set the claims in context, so handlers can get it.
		c.Set(ContextKeyClaims, claims)
		c.Next()
	}
}
