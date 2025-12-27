package handlers

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/libpulse/platform/services/api/internal/auth"
	"github.com/libpulse/platform/services/api/internal/supabase"
	"github.com/libpulse/platform/services/api/internal/utils/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Initialize crypto package for all handler tests
func init() {
	crypto.Init("test-pepper-for-handler-tests")
}

// MockProjectStore implements handlers.ProjectStore for testing.
type MockProjectStore struct {
	mock.Mock
}

// NewMockProjectStore creates a new mock ProjectStore.
func NewMockProjectStore() *MockProjectStore {
	return &MockProjectStore{}
}

// GetProjectByID mocks ProjectStore.GetProjectByID.
func (m *MockProjectStore) GetProjectByID(
	ctx context.Context,
	projectID string,
) (*supabase.Project, error) {
	args := m.Called(ctx, projectID)

	var project *supabase.Project
	if v := args.Get(0); v != nil {
		project = v.(*supabase.Project)
	}

	return project, args.Error(1)
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

// MockProjectKeyStore implements handlers.ProjectKeyStore for testing
type MockProjectKeyStore struct {
	mock.Mock
}

func NewMockProjectKeyStore() *MockProjectKeyStore {
	return &MockProjectKeyStore{}
}

func (m *MockProjectKeyStore) CreateProjectKey(
	ctx context.Context,
	params supabase.CreateProjectKeyParams,
) (*supabase.ProjectKey, error) {
	args := m.Called(ctx, params)

	var key *supabase.ProjectKey
	if v := args.Get(0); v != nil {
		key = v.(*supabase.ProjectKey)
	}

	return key, args.Error(1)
}

// TestCreateProjectKeyHandler_Success tests successful project key creation
func TestCreateProjectKeyHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockProjectStore := NewMockProjectStore()
	mockKeyStore := NewMockProjectKeyStore()

	userID := "user-success-test"
	projectID := "proj-456"

	project := &supabase.Project{
		ID:          projectID,
		Name:        "test-project",
		OwnerUserID: userID,
	}

	// Mock project lookup
	mockProjectStore.On("GetProjectByID", mock.Anything, projectID).Return(project, nil)

	// Mock key creation - match any params since we can't predict generated keys
	keyData := &supabase.ProjectKey{
		ID:                "key-789",
		ProjectID:         projectID,
		Label:             "test-key",
		Env:               "prod",
		SignedOnly:        false,
		PublicKey:         "pk_live_test",
		SecretEnc:         "hash",
		SecretFingerprint: "1234",
		CreatedBy:         userID,
		CreatedAt:         time.Now(),
	}
	mockKeyStore.On("CreateProjectKey", mock.Anything, mock.AnythingOfType("supabase.CreateProjectKeyParams")).Return(keyData, nil)

	// Create HTTP test context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: projectID}}

	requestBody := `{"label":"test-key"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+projectID+"/keys", bytes.NewBufferString(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	claims := &auth.SupabaseClaims{
		Email: "test@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userID,
		},
	}
	c.Set(auth.ContextKeyClaims, claims)

	// Execute handler
	handler := CreateProjectKeyHandler(mockProjectStore, mockKeyStore)
	handler(c)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "project_key_public")
	assert.Contains(t, w.Body.String(), "project_secret")
	assert.Contains(t, w.Body.String(), "pk_live_")
	assert.Contains(t, w.Body.String(), "psk_live_")
	assert.Equal(t, "no-store", w.Header().Get("Cache-Control"))

	mockProjectStore.AssertExpectations(t)
	mockKeyStore.AssertExpectations(t)
}

// TestCreateProjectKeyHandler_NoAuth tests missing authentication
func TestCreateProjectKeyHandler_NoAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockProjectStore := NewMockProjectStore()
	mockKeyStore := NewMockProjectKeyStore()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "proj-123"}}

	requestBody := `{"label":"test-key"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/projects/proj-123/keys", bytes.NewBufferString(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler := CreateProjectKeyHandler(mockProjectStore, mockKeyStore)
	handler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockProjectStore.AssertNotCalled(t, "GetProjectByID")
	mockKeyStore.AssertNotCalled(t, "CreateProjectKey")
}

// TestCreateProjectKeyHandler_InvalidClaimsType tests invalid claims type
func TestCreateProjectKeyHandler_InvalidClaimsType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockProjectStore := NewMockProjectStore()
	mockKeyStore := NewMockProjectKeyStore()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "proj-123"}}
	c.Set(auth.ContextKeyClaims, "invalid-type")

	requestBody := `{"label":"test-key"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/projects/proj-123/keys", bytes.NewBufferString(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler := CreateProjectKeyHandler(mockProjectStore, mockKeyStore)
	handler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestCreateProjectKeyHandler_MissingProjectID tests missing project ID in URL
func TestCreateProjectKeyHandler_MissingProjectID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockProjectStore := NewMockProjectStore()
	mockKeyStore := NewMockProjectKeyStore()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// No project ID param

	claims := &auth.SupabaseClaims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: "user-123"},
	}
	c.Set(auth.ContextKeyClaims, claims)

	requestBody := `{"label":"test-key"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/projects//keys", bytes.NewBufferString(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler := CreateProjectKeyHandler(mockProjectStore, mockKeyStore)
	handler(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestCreateProjectKeyHandler_InvalidRequestBody tests invalid JSON
func TestCreateProjectKeyHandler_InvalidRequestBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockProjectStore := NewMockProjectStore()
	mockKeyStore := NewMockProjectKeyStore()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "proj-123"}}

	claims := &auth.SupabaseClaims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: "user-123"},
	}
	c.Set(auth.ContextKeyClaims, claims)

	requestBody := `{"invalid json`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/projects/proj-123/keys", bytes.NewBufferString(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler := CreateProjectKeyHandler(mockProjectStore, mockKeyStore)
	handler(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestCreateProjectKeyHandler_MissingLabel tests missing required label field
func TestCreateProjectKeyHandler_MissingLabel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockProjectStore := NewMockProjectStore()
	mockKeyStore := NewMockProjectKeyStore()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "proj-123"}}

	claims := &auth.SupabaseClaims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: "user-123"},
	}
	c.Set(auth.ContextKeyClaims, claims)

	requestBody := `{}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/projects/proj-123/keys", bytes.NewBufferString(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler := CreateProjectKeyHandler(mockProjectStore, mockKeyStore)
	handler(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestCreateProjectKeyHandler_ProjectNotFound tests non-existent project
func TestCreateProjectKeyHandler_ProjectNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockProjectStore := NewMockProjectStore()
	mockKeyStore := NewMockProjectKeyStore()

	projectID := "proj-nonexistent"
	userID := "user-notfound-test"

	mockProjectStore.On("GetProjectByID", mock.Anything, projectID).Return(nil, errors.New("project not found"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: projectID}}

	claims := &auth.SupabaseClaims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: userID},
	}
	c.Set(auth.ContextKeyClaims, claims)

	requestBody := `{"label":"test-key"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+projectID+"/keys", bytes.NewBufferString(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler := CreateProjectKeyHandler(mockProjectStore, mockKeyStore)
	handler(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "not_found")
	mockProjectStore.AssertExpectations(t)
	mockKeyStore.AssertNotCalled(t, "CreateProjectKey")
}

// TestCreateProjectKeyHandler_InvalidProjectID tests invalid project ID parameter
func TestCreateProjectKeyHandler_InvalidProjectID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockProjectStore := NewMockProjectStore()
	mockKeyStore := NewMockProjectKeyStore()

	projectID := "invalid-uuid"
	userID := "user-invalidid-test"

	// Mock database returning error for invalid UUID format
	mockProjectStore.On("GetProjectByID", mock.Anything, projectID).Return(nil, errors.New("invalid input syntax for type uuid"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: projectID}}

	claims := &auth.SupabaseClaims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: userID},
	}
	c.Set(auth.ContextKeyClaims, claims)

	requestBody := `{"label":"test-key"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+projectID+"/keys", bytes.NewBufferString(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler := CreateProjectKeyHandler(mockProjectStore, mockKeyStore)
	handler(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "bad_request")
	mockProjectStore.AssertExpectations(t)
	mockKeyStore.AssertNotCalled(t, "CreateProjectKey")
}

// TestCreateProjectKeyHandler_NotOwner tests user who is not the project owner
func TestCreateProjectKeyHandler_NotOwner(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockProjectStore := NewMockProjectStore()
	mockKeyStore := NewMockProjectKeyStore()

	projectID := "proj-456"
	ownerID := "owner-123"
	otherUserID := "user-999"

	project := &supabase.Project{
		ID:          projectID,
		Name:        "test-project",
		OwnerUserID: ownerID, // Different from requesting user
	}

	mockProjectStore.On("GetProjectByID", mock.Anything, projectID).Return(project, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: projectID}}

	claims := &auth.SupabaseClaims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: otherUserID},
	}
	c.Set(auth.ContextKeyClaims, claims)

	requestBody := `{"label":"test-key"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+projectID+"/keys", bytes.NewBufferString(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler := CreateProjectKeyHandler(mockProjectStore, mockKeyStore)
	handler(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "forbidden")
	mockProjectStore.AssertExpectations(t)
	mockKeyStore.AssertNotCalled(t, "CreateProjectKey")
}

// TestCreateProjectKeyHandler_DatabaseError tests database error during key creation
func TestCreateProjectKeyHandler_DatabaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockProjectStore := NewMockProjectStore()
	mockKeyStore := NewMockProjectKeyStore()

	userID := "user-dberror-test"
	projectID := "proj-456"

	project := &supabase.Project{
		ID:          projectID,
		Name:        "test-project",
		OwnerUserID: userID,
	}

	mockProjectStore.On("GetProjectByID", mock.Anything, projectID).Return(project, nil)
	mockKeyStore.On("CreateProjectKey", mock.Anything, mock.AnythingOfType("supabase.CreateProjectKeyParams")).Return(nil, errors.New("database error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: projectID}}

	claims := &auth.SupabaseClaims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: userID},
	}
	c.Set(auth.ContextKeyClaims, claims)

	requestBody := `{"label":"test-key"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+projectID+"/keys", bytes.NewBufferString(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler := CreateProjectKeyHandler(mockProjectStore, mockKeyStore)
	handler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "internal_error")
	mockProjectStore.AssertExpectations(t)
	mockKeyStore.AssertExpectations(t)
}

// TestCreateProjectKeyHandler_WithCustomEnv tests custom env parameter
func TestCreateProjectKeyHandler_WithCustomEnv(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockProjectStore := NewMockProjectStore()
	mockKeyStore := NewMockProjectKeyStore()

	userID := "user-customenv-test"
	projectID := "proj-456"

	project := &supabase.Project{
		ID:          projectID,
		Name:        "test-project",
		OwnerUserID: userID,
	}

	mockProjectStore.On("GetProjectByID", mock.Anything, projectID).Return(project, nil)

	keyData := &supabase.ProjectKey{
		ID:                "key-789",
		ProjectID:         projectID,
		Label:             "staging-key",
		Env:               "staging",
		CreatedBy:         userID,
		CreatedAt:         time.Now(),
	}
	mockKeyStore.On("CreateProjectKey", mock.Anything, mock.AnythingOfType("supabase.CreateProjectKeyParams")).Return(keyData, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: projectID}}

	claims := &auth.SupabaseClaims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: userID},
	}
	c.Set(auth.ContextKeyClaims, claims)

	requestBody := `{"label":"staging-key","env":"staging","require_signature":true}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+projectID+"/keys", bytes.NewBufferString(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler := CreateProjectKeyHandler(mockProjectStore, mockKeyStore)
	handler(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockProjectStore.AssertExpectations(t)
	mockKeyStore.AssertExpectations(t)
}
