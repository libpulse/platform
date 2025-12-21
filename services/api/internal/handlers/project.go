package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/libpulse/platform/services/api/internal/auth"
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
