package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/libpulse/platform/services/api/internal/auth"
	"github.com/libpulse/platform/services/api/internal/utils/errors"
)

// Get User info when authenticated: /api/v1/me
func GetCurrentUserHandler(store UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1) ensure claims(injected by auth middleware)
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

		// 3) Get user info
		user, err := store.GetUserByID(c.Request.Context(), claims.Subject)
		if err != nil {
			apiErr := errors.NewAPIError(errors.ErrInternalError)
			c.JSON(apiErr.StatusCode(), apiErr)
			return
		}

		// 4） Defensive: store returned (nil, nil), which should never happen.
		if user == nil {
			apiErr := errors.NewAPIError(errors.ErrInternalError)
			c.JSON(apiErr.StatusCode(), apiErr)
			return
		}

		c.JSON(http.StatusOK, user)
	}
}
