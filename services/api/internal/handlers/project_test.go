package handlers

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/libpulse/platform/services/api/internal/auth"
	"github.com/libpulse/platform/services/api/internal/supabase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProjectStore implements handlers.ProjectStore for testing.
type MockProjectStore struct {
	mock.Mock
}

// NewMockProjectStore creates a new mock ProjectStore.
func NewMockProjectStore() *MockProjectStore {
	return &MockProjectStore{}
}

// CreateProject mocks ProjectStore.CreateProject.
func (m *MockProjectStore) CreateProject(
	ctx context.Context,
	name string,
	ownerUserID string,
) (*supabase.Project, error) {
	args := m.Called(ctx, name, ownerUserID)

	var project *supabase.Project
	if v := args.Get(0); v != nil {
		project = v.(*supabase.Project)
	}

	return project, args.Error(1)
}

// TestCreateProjectHandler_Success tests successful project creation
func TestCreateProjectHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := NewMockProjectStore()

	userID := "user-123-abc"
	projectID := "project-456-def"
	projectName := "demo"

	projectData := &supabase.Project{
		ID:          projectID,
		Name:        projectName,
		OwnerUserID: userID,
	}

	// Mock Supabase response
	mockStore.On("CreateProject", mock.Anything, projectName, userID).Return(projectData, nil)

	// Create HTTP test context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	requestBody := `{"name":"demo"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewBufferString(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Create claims with Subject field
	claims := &auth.SupabaseClaims{
		Email: "test@example.com",
		Role:  "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userID,
		},
	}
	c.Set(auth.ContextKeyClaims, claims)

	// Execute handler
	handler := CreateProjectHandler(mockStore)
	handler(c)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), projectID)
	mockStore.AssertExpectations(t)
}

// TestCreateProjectHandler_DatabaseError tests error handling
func TestCreateProjectHandler_DatabaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := NewMockProjectStore()

	userID := "user-123-abc"
	projectName := "demo"

	// Mock Supabase error
	mockStore.On("CreateProject", mock.Anything, projectName, userID).Return(nil, errors.New("database error"))

	// Create HTTP test context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	requestBody := `{"name":"demo"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewBufferString(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	claims := &auth.SupabaseClaims{
		Email: "test@example.com",
		Role:  "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userID,
		},
	}
	c.Set(auth.ContextKeyClaims, claims)

	// Execute handler
	handler := CreateProjectHandler(mockStore)
	handler(c)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "internal_error")
	mockStore.AssertExpectations(t)
}

// TestCreateProjectHandler_NoClaimsInContext tests when claims aren't set
func TestCreateProjectHandler_NoClaimsInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := NewMockProjectStore()

	// Create HTTP test context WITHOUT setting claims
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	requestBody := `{"name":"demo"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewBufferString(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler := CreateProjectHandler(mockStore)
	handler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")
	mockStore.AssertNotCalled(t, "CreateProject", mock.Anything, mock.Anything, mock.Anything)
}

// TestCreateProjectHandler_EmptyUserID tests with empty user ID
func TestCreateProjectHandler_EmptyUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := NewMockProjectStore()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	requestBody := `{"name":"demo"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewBufferString(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Create claims with empty user id
	claims := &auth.SupabaseClaims{
		Email: "test@example.com",
		Role:  "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "",
		},
	}
	c.Set(auth.ContextKeyClaims, claims)

	handler := CreateProjectHandler(mockStore)
	handler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockStore.AssertNotCalled(t, "CreateProject", mock.Anything, mock.Anything, mock.Anything)
}

// TestCreateProjectHandler_InvalidClaimsType tests the case where an invalid claims type exists
func TestCreateProjectHandler_InvalidClaimsType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := NewMockProjectStore()

	// Create HTTP test context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	requestBody := `{"name":"demo"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewBufferString(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Put a wrong type into context under the claims key
	c.Set(auth.ContextKeyClaims, "not-a-claims-struct")

	// Execute handler
	handler := CreateProjectHandler(mockStore)
	handler(c)

	// Assertions: should be unauthorized
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")

	// Ensure store isn't called
	mockStore.AssertNotCalled(t, "CreateProject", mock.Anything, mock.Anything, mock.Anything)
	mockStore.AssertExpectations(t)
}

// TestCreateProjectHandler_InvalidRequestBody tests invalid request body
func TestCreateProjectHandler_InvalidRequestBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := NewMockProjectStore()

	userID := "user-123-abc"

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Invalid JSON body
	requestBody := `{"invalid": "json"`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewBufferString(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	claims := &auth.SupabaseClaims{
		Email: "test@example.com",
		Role:  "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userID,
		},
	}
	c.Set(auth.ContextKeyClaims, claims)

	handler := CreateProjectHandler(mockStore)
	handler(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "bad_request")
	mockStore.AssertNotCalled(t, "CreateProject", mock.Anything, mock.Anything, mock.Anything)
}

// TestCreateProjectHandler_MissingName tests missing required name field
func TestCreateProjectHandler_MissingName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := NewMockProjectStore()

	userID := "user-123-abc"

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Missing name field
	requestBody := `{}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewBufferString(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	claims := &auth.SupabaseClaims{
		Email: "test@example.com",
		Role:  "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userID,
		},
	}
	c.Set(auth.ContextKeyClaims, claims)

	handler := CreateProjectHandler(mockStore)
	handler(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "bad_request")
	mockStore.AssertNotCalled(t, "CreateProject", mock.Anything, mock.Anything, mock.Anything)
}

// TestCreateProjectHandler_EmptyName tests empty name field
func TestCreateProjectHandler_EmptyName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := NewMockProjectStore()

	userID := "user-123-abc"

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Empty name
	requestBody := `{"name":""}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewBufferString(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	claims := &auth.SupabaseClaims{
		Email: "test@example.com",
		Role:  "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userID,
		},
	}
	c.Set(auth.ContextKeyClaims, claims)

	handler := CreateProjectHandler(mockStore)
	handler(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "bad_request")
	mockStore.AssertNotCalled(t, "CreateProject", mock.Anything, mock.Anything, mock.Anything)
}

// TestCreateProjectHandler_NilProjectNoError tests the case where the store returns (nil, nil)
func TestCreateProjectHandler_NilProjectNoError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := NewMockProjectStore()

	userID := "user-123-abc"
	projectName := "demo"

	// Create HTTP test context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	requestBody := `{"name":"demo"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewBufferString(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Set valid claims
	claims := &auth.SupabaseClaims{
		Email: "test@example.com",
		Role:  "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userID,
		},
	}
	c.Set(auth.ContextKeyClaims, claims)

	// Mock store returning (nil, nil) — contract violation
	mockStore.On("CreateProject", c.Request.Context(), projectName, userID).
		Return((*supabase.Project)(nil), nil)

	// Execute handler
	handler := CreateProjectHandler(mockStore)
	handler(c)

	// Should be treated as internal error
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "internal_error")

	mockStore.AssertExpectations(t)
}

// TestCreateProjectHandler_DuplicateProjectName tests duplicate project name conflict
func TestCreateProjectHandler_DuplicateProjectName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := NewMockProjectStore()

	userID := "user-123-abc"
	projectName := "demo"

	// Mock Supabase error for duplicate key constraint
	duplicateErr := errors.New("duplicate key value violates unique constraint \"projects_owner_name_unique\"")
	mockStore.On("CreateProject", mock.Anything, projectName, userID).Return(nil, duplicateErr)

	// Create HTTP test context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	requestBody := `{"name":"demo"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewBufferString(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	claims := &auth.SupabaseClaims{
		Email: "test@example.com",
		Role:  "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userID,
		},
	}
	c.Set(auth.ContextKeyClaims, claims)

	// Execute handler
	handler := CreateProjectHandler(mockStore)
	handler(c)

	// Assertions: should return 409 Conflict
	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "conflict")
	mockStore.AssertExpectations(t)
}
