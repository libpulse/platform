package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/libpulse/platform/services/api/internal/auth"
	"github.com/libpulse/platform/services/api/internal/supabase"
)

type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

// Get User info when authenticated: /api/v1/me
func GetCurrentUserHandler(sb *supabase.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		claimsAny, ok := c.Get(auth.ContextKeyClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error: "no auth context",
				Code:  "unauthorized",
			})
			return
		}

		claims, ok := claimsAny.(*auth.SupabaseClaims)
		if !ok || claims.Subject == "" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error: "invalid auth context",
				Code:  "unauthorized",
			})
			return
		}

		user, err := sb.GetUserByID(c.Request.Context(), claims.Subject)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: "failed to fetch user",
				Code:  "internal_error",
			})
			return
		}

		c.JSON(http.StatusOK, user)
	}
}
