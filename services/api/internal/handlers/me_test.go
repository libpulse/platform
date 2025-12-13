package handlers

import (
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

// MockUserStore implements handlers.UserStore for testing.
type MockUserStore struct {
	mock.Mock
}

// NewMockUserStore creates a new mock UserStore.
func NewMockUserStore() *MockUserStore {
	return &MockUserStore{}
}

// GetUserByID mocks UserStore.GetUserByID.
func (m *MockUserStore) GetUserByID(
	ctx context.Context,
	userID string,
) (*supabase.User, error) {
	args := m.Called(ctx, userID)

	var user *supabase.User
	if v := args.Get(0); v != nil {
		user = v.(*supabase.User)
	}

	return user, args.Error(1)
}

// TestGetCurrentUserHandler_Success tests successful user retrieval
func TestGetCurrentUserHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := NewMockUserStore()

	userID := "user-123-abc"
	userData := &supabase.User{
		ID:    userID,
		Email: "test@example.com",
	}

	// Mock Supabase response
	mockStore.On("GetUserByID", mock.Anything, userID).Return(userData, nil)

	// Create HTTP test context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)

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
	handler := GetCurrentUserHandler(mockStore)
	handler(c)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "test@example.com")
	mockStore.AssertExpectations(t)
}

// TestGetCurrentUserHandler_DatabaseError tests error handling
func TestGetCurrentUserHandler_DatabaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := NewMockUserStore()

	userID := "user-123-abc"

	// Mock Supabase error
	mockStore.On("GetUserByID", mock.Anything, userID).Return(nil, errors.New("database error"))

	// Create HTTP test context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)

	claims := &auth.SupabaseClaims{
		Email: "test@example.com",
		Role:  "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userID,
		},
	}
	c.Set(auth.ContextKeyClaims, claims)

	// Execute handler
	handler := GetCurrentUserHandler(mockStore)
	handler(c)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "internal_error")
	mockStore.AssertExpectations(t)
}

// TestGetCurrentUserHandler_NoClaimsInContext tests when claims aren't set
func TestGetCurrentUserHandler_NoClaimsInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := NewMockUserStore()

	// Create HTTP test context WITHOUT setting claims
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler := GetCurrentUserHandler(mockStore)
	handler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")
	mockStore.AssertNotCalled(t, "GetUserByID", mock.Anything, mock.Anything)
}

// TestGetCurrentUserHandler_EmptyUserID tests with empty user ID
func TestGetCurrentUserHandler_EmptyUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := NewMockUserStore()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)

	// Create claims with empty user id.
	claims := &auth.SupabaseClaims{
		Email: "test@example.com",
		Role:  "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "",
		},
	}
	c.Set(auth.ContextKeyClaims, claims)

	handler := GetCurrentUserHandler(mockStore)
	handler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockStore.AssertNotCalled(t, "GetUserByID", mock.Anything, mock.Anything)
}

// TestGetCurrentUserHandler_InvalidClaimsType tests the case where an invalid claims type exists;
// This should return 401.
func TestGetCurrentUserHandler_InvalidClaimsType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := NewMockUserStore()

	// Create HTTP test context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)

	// Put a wrong type into context under the claims key
	c.Set(auth.ContextKeyClaims, "not-a-claims-struct")

	// Execute handler
	handler := GetCurrentUserHandler(mockStore)
	handler(c)

	// Assertions: should be unauthorized
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")

	// Ensure store isn't called
	mockStore.AssertNotCalled(t, "GetUserByID", mock.Anything, mock.Anything)
	mockStore.AssertExpectations(t)
}

// TestGetCurrentUserHandler_NilUserNoError tests the case where the store returns (nil, nil).
// This should be treated as an internal error (contract violation).
func TestGetCurrentUserHandler_NilUserNoError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := NewMockUserStore()

	userID := "user-123-abc"

	// Create HTTP test context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)

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
	mockStore.On("GetUserByID", c.Request.Context(), userID).
		Return((*supabase.User)(nil), nil)

	// Execute handler
	handler := GetCurrentUserHandler(mockStore)
	handler(c)

	// What SHOULD happen:
	// returning (nil, nil) should be treated as internal error
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "internal_error")

	mockStore.AssertExpectations(t)
}
