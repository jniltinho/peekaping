package api_key

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockService implements Service interface for middleware tests
type MockService struct {
	mock.Mock
}

func (m *MockService) Create(ctx context.Context, req *CreateRequest) (*APIKeyWithToken, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*APIKeyWithToken), args.Error(1)
}

func (m *MockService) FindAll(ctx context.Context) ([]*Model, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Model), args.Error(1)
}

func (m *MockService) FindByID(ctx context.Context, id string) (*Model, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Model), args.Error(1)
}

func (m *MockService) Update(ctx context.Context, id string, req *UpdateRequest) (*Model, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Model), args.Error(1)
}

func (m *MockService) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockService) ValidateKey(ctx context.Context, key string) (*Model, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Model), args.Error(1)
}

func TestMiddlewareProvider_Auth_MissingHeader(t *testing.T) {
	mockSvc := new(MockService)
	middleware := NewMiddlewareProvider(mockSvc)

	e := echo.New()
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// No X-API-Key header set

	next := func(c *echo.Context) error { return nil }
	err := middleware.Auth()(next)(c)

	// In our port, it returns JSON error directly
	assert.Error(t, err)

	// Check response (status written by c.JSON inside middleware)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp map[string]interface{}
	jsonErr := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, jsonErr)
	assert.Contains(t, resp["message"], "X-API-Key header is required")
	assert.Nil(t, resp["data"])

	mockSvc.AssertExpectations(t)
}

func TestMiddlewareProvider_Auth_InvalidPrefix(t *testing.T) {
	mockSvc := new(MockService)
	middleware := NewMiddlewareProvider(mockSvc)

	e := echo.New()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-API-Key", "invalid-prefix-token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	next := func(c *echo.Context) error { return nil }
	err := middleware.Auth()(next)(c)

	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	// Check response body directly
	var resp map[string]interface{}
	jsonErr := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, jsonErr)
	assert.Contains(t, resp["message"], "API key required")
	assert.Nil(t, resp["data"])

	mockSvc.AssertExpectations(t)
}

func TestMiddlewareProvider_Auth_ValidKey(t *testing.T) {
	mockSvc := new(MockService)
	middleware := NewMiddlewareProvider(mockSvc)

	validToken := ApiKeyPrefix + "eyJpZCI6InRlc3Qta2V5LWlkIiwia2V5IjoidGVzdC1hY3R1YWwta2V5In0="
	apiKey := &Model{
		ID:         "test-key-id",
		Name:       "Test Key",
		DisplayKey: "pk_test...",
	}

	mockSvc.On("ValidateKey", mock.Anything, validToken).Return(apiKey, nil)

	e := echo.New()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-API-Key", validToken)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	next := func(c *echo.Context) error { return nil }
	err := middleware.Auth()(next)(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify context values are set
	apiKeyId := c.Get("apiKeyId")
	assert.Equal(t, "test-key-id", apiKeyId)

	authType := c.Get("authType")
	assert.Equal(t, "api_key", authType)

	mockSvc.AssertExpectations(t)
}

func TestMiddlewareProvider_Auth_ExpiredKey(t *testing.T) {
	mockSvc := new(MockService)
	middleware := NewMiddlewareProvider(mockSvc)

	expiredToken := ApiKeyPrefix + "eyJpZCI6ImV4cGlyZWQta2V5LWlkIiwia2V5IjoidGVzdC1hY3R1YWwta2V5In0="

	mockSvc.On("ValidateKey", mock.Anything, expiredToken).Return(nil, errors.New("API key has expired"))

	e := echo.New()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-API-Key", expiredToken)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	next := func(c *echo.Context) error { return nil }
	err := middleware.Auth()(next)(c)

	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	// Check response body directly
	var resp map[string]interface{}
	jsonErr := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, jsonErr)
	assert.Contains(t, resp["message"], "Invalid or expired API key")
	assert.Nil(t, resp["data"])

	mockSvc.AssertExpectations(t)
}

func TestMiddlewareProvider_Auth_InvalidKey(t *testing.T) {
	mockSvc := new(MockService)
	middleware := NewMiddlewareProvider(mockSvc)

	invalidToken := ApiKeyPrefix + "eyJpZCI6ImludmFsaWQta2V5LWlkIiwia2V5IjoidGVzdC1hY3R1YWwta2V5In0="

	mockSvc.On("ValidateKey", mock.Anything, invalidToken).Return(nil, errors.New("Invalid API key"))

	e := echo.New()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-API-Key", invalidToken)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	next := func(c *echo.Context) error { return nil }
	err := middleware.Auth()(next)(c)

	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	// Check response body directly
	var resp map[string]interface{}
	jsonErr := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, jsonErr)
	assert.Contains(t, resp["message"], "Invalid or expired API key")
	assert.Nil(t, resp["data"])

	mockSvc.AssertExpectations(t)
}

func TestMiddlewareProvider_Auth_ContextValues(t *testing.T) {
	mockSvc := new(MockService)
	middleware := NewMiddlewareProvider(mockSvc)

	validToken := ApiKeyPrefix + "eyJpZCI6InRlc3Qta2V5LWlkIiwia2V5IjoidGVzdC1hY3R1YWwta2V5In0="
	apiKey := &Model{
		ID:         "context-test-key",
		Name:       "Context Test Key",
		DisplayKey: "pk_context...",
	}

	mockSvc.On("ValidateKey", mock.Anything, validToken).Return(apiKey, nil)

	e := echo.New()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-API-Key", validToken)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	next := func(c *echo.Context) error { return nil }
	err := middleware.Auth()(next)(c)

	assert.NoError(t, err)

	// Verify specific context values
	apiKeyId := c.Get("apiKeyId")
	assert.Equal(t, "context-test-key", apiKeyId)

	authType := c.Get("authType")
	assert.Equal(t, "api_key", authType)

	// Verify no other auth-related context values are set (Get returns nil if not present)
	jwtUserId := c.Get("userId")
	assert.Nil(t, jwtUserId)

	mockSvc.AssertExpectations(t)
}

func TestMiddlewareProvider_Auth_ServiceError(t *testing.T) {
	mockSvc := new(MockService)
	middleware := NewMiddlewareProvider(mockSvc)

	validToken := ApiKeyPrefix + "eyJpZCI6InNlcnZpY2UtZXJyb3Ita2V5Iiwia2V5IjoidGVzdC1hY3R1YWwta2V5In0="

	mockSvc.On("ValidateKey", mock.Anything, validToken).Return(nil, errors.New("service error"))

	e := echo.New()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-API-Key", validToken)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	next := func(c *echo.Context) error { return nil }
	err := middleware.Auth()(next)(c)

	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	// Check response body directly
	var resp map[string]interface{}
	jsonErr := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, jsonErr)
	assert.Contains(t, resp["message"], "Invalid or expired API key")
	assert.Nil(t, resp["data"])

	mockSvc.AssertExpectations(t)
}
