package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/libpulse/platform/services/api/internal/auth"
	"github.com/libpulse/platform/services/api/internal/supabase"
	"github.com/libpulse/platform/services/api/internal/utils/errors"
)

type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

// Get User info when authenticated: /api/v1/me
func GetCurrentUserHandler(sb *supabase.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get claims from context (set by auth middleware)
		claimsAny, _ := c.Get(auth.ContextKeyClaims)
		claims := claimsAny.(*auth.SupabaseClaims)

		user, err := sb.GetUserByID(c.Request.Context(), claims.Subject)
		if err != nil {
			apiErr := errors.NewAPIError(errors.ErrInternalError)
			c.JSON(apiErr.StatusCode(), apiErr)
			return
		}

		c.JSON(http.StatusOK, user)
	}
}
