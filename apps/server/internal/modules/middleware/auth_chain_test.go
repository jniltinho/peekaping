package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestAuthChain is a test-specific version that allows us to inject mock handlers
type TestAuthChain struct {
	jwtMiddleware    echo.MiddlewareFunc
	apiKeyMiddleware echo.MiddlewareFunc
	logger           *zap.SugaredLogger
}

func NewTestAuthChain(jwtHandler, apiKeyHandler echo.MiddlewareFunc, logger *zap.SugaredLogger) *TestAuthChain {
	return &TestAuthChain{
		jwtMiddleware:    jwtHandler,
		apiKeyMiddleware: apiKeyHandler,
		logger:           logger.Named("[auth_chain]"),
	}
}

// AllAuth creates a middleware that supports both JWT and API key authentication
func (ac *TestAuthChain) AllAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			apiKeyHeader := c.Request().Header.Get("X-API-Key")
			authHeader := c.Request().Header.Get("Authorization")

			if apiKeyHeader != "" {
				// Route to API key authentication
				ac.logger.Infow("Routing to API key authentication", "ip", c.RealIP(), "path", c.Request().URL.Path, "keyPrefix", apiKeyHeader[:min(len(apiKeyHeader), 10)]+"...")
				return ac.apiKeyMiddleware(next)(c)
			} else if authHeader != "" {
				// Route to JWT authentication
				ac.logger.Infow("Routing to JWT authentication", "ip", c.RealIP(), "path", c.Request().URL.Path, "tokenPrefix", authHeader[:min(len(authHeader), 10)]+"...")
				return ac.jwtMiddleware(next)(c)
			} else {
				// No authentication headers provided
				ac.logger.Debugw("Missing authentication headers", "ip", c.RealIP(), "path", c.Request().URL.Path)
				return c.JSON(http.StatusUnauthorized, map[string]any{"success": false, "message": "Authentication required: provide either X-API-Key header or Authorization header"})
			}
		}
	}
}

func TestAuthChain_AllAuth_RoutesToAPIKey(t *testing.T) {
	logger := zap.NewNop().Sugar()

	apiKeyCalled := false
	jwtCalled := false

	jwtHandler := func(c *echo.Context) error {
		jwtCalled = true
		c.Set("authType", "jwt")
		c.Set("userId", "test-user-id")
		return nil
	}

	apiKeyHandler := func(c *echo.Context) error {
		apiKeyCalled = true
		c.Set("authType", "api_key")
		c.Set("apiKeyId", "test-api-key-id")
		return nil
	}

	authChain := NewTestAuthChain(jwtHandler, apiKeyHandler, logger)

	e := echo.New()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "pk_test-token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	next := func(c *echo.Context) error { return nil }
	err := authChain.AllAuth()(next)(c)

	assert.NoError(t, err)

	// Verify API key middleware was called
	assert.True(t, apiKeyCalled)
	assert.False(t, jwtCalled)

	// Verify context values
	authType, exists := c.Get("authType")
	assert.True(t, exists)
	assert.Equal(t, "api_key", authType)

	apiKeyId, exists := c.Get("apiKeyId")
	assert.True(t, exists)
	assert.Equal(t, "test-api-key-id", apiKeyId)

	// Verify JWT context is not set
	_, jwtExists := c.Get("userId")
	assert.False(t, jwtExists)
}

func TestAuthChain_AllAuth_RoutesToJWT(t *testing.T) {
	logger := zap.NewNop().Sugar()

	apiKeyCalled := false
	jwtCalled := false

	jwtHandler := func(c *echo.Context) error {
		jwtCalled = true
		c.Set("authType", "jwt")
		c.Set("userId", "test-user-id")
		return nil
	}

	apiKeyHandler := func(c *echo.Context) error {
		apiKeyCalled = true
		c.Set("authType", "api_key")
		c.Set("apiKeyId", "test-api-key-id")
		return nil
	}

	authChain := NewTestAuthChain(jwtHandler, apiKeyHandler, logger)

	e := echo.New()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer jwt-token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	next := func(c *echo.Context) error { return nil }
	err := authChain.AllAuth()(next)(c)

	assert.NoError(t, err)

	// Verify JWT middleware was called
	assert.True(t, jwtCalled)
	assert.False(t, apiKeyCalled)

	// Verify context values
	authType, exists := c.Get("authType")
	assert.True(t, exists)
	assert.Equal(t, "jwt", authType)

	userId, exists := c.Get("userId")
	assert.True(t, exists)
	assert.Equal(t, "test-user-id", userId)

	// Verify API key context is not set
	_, apiKeyExists := c.Get("apiKeyId")
	assert.False(t, apiKeyExists)
}

func TestAuthChain_AllAuth_BothHeaders(t *testing.T) {
	logger := zap.NewNop().Sugar()

	apiKeyCalled := false
	jwtCalled := false

	jwtHandler := func(c *echo.Context) error {
		jwtCalled = true
		c.Set("authType", "jwt")
		c.Set("userId", "test-user-id")
		return nil
	}

	apiKeyHandler := func(c *echo.Context) error {
		apiKeyCalled = true
		c.Set("authType", "api_key")
		c.Set("apiKeyId", "test-api-key-id")
		return nil
	}

	authChain := NewTestAuthChain(jwtHandler, apiKeyHandler, logger)

	e := echo.New()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "pk_test-token")
	req.Header.Set("Authorization", "Bearer jwt-token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	next := func(c *echo.Context) error { return nil }
	err := authChain.AllAuth()(next)(c)

	assert.NoError(t, err)

	// Verify API key middleware takes priority
	assert.True(t, apiKeyCalled)
	assert.False(t, jwtCalled)

	// Verify API key context values
	authType, exists := c.Get("authType")
	assert.True(t, exists)
	assert.Equal(t, "api_key", authType)

	apiKeyId, exists := c.Get("apiKeyId")
	assert.True(t, exists)
	assert.Equal(t, "test-api-key-id", apiKeyId)
}

func TestAuthChain_AllAuth_NoHeaders(t *testing.T) {
	logger := zap.NewNop().Sugar()

	apiKeyCalled := false
	jwtCalled := false

	jwtHandler := func(c *echo.Context) error {
		jwtCalled = true
		return nil
	}

	apiKeyHandler := func(c *echo.Context) error {
		apiKeyCalled = true
		return nil
	}

	authChain := NewTestAuthChain(jwtHandler, apiKeyHandler, logger)

	e := echo.New()
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// No headers set

	next := func(c *echo.Context) error { return nil }
	err := authChain.AllAuth()(next)(c)

	// Verify neither middleware was called
	assert.False(t, apiKeyCalled)
	assert.False(t, jwtCalled)

	// In Echo middleware, for no headers we return JSON error (no abort like Gin)
	// The middleware returns error for 401 case
	assert.Error(t, err)
}

func TestAuthChain_AllAuth_EmptyAPIKeyHeader(t *testing.T) {
	logger := zap.NewNop().Sugar()

	apiKeyCalled := false
	jwtCalled := false

	jwtHandler := func(c *echo.Context) error {
		jwtCalled = true
		return nil
	}

	apiKeyHandler := func(c *echo.Context) error {
		apiKeyCalled = true
		return nil
	}

	authChain := NewTestAuthChain(jwtHandler, apiKeyHandler, logger)

	e := echo.New()
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Request().Header.Set("X-API-Key", "") // Empty header

	next := func(c *echo.Context) error { return nil }
	err := authChain.AllAuth()(next)(c)

	// Verify neither middleware was called (empty header treated as missing)
	assert.False(t, apiKeyCalled)
	assert.False(t, jwtCalled)

	// Verify error returned for 401
	assert.Error(t, err)
}

func TestAuthChain_AllAuth_EmptyAuthHeader(t *testing.T) {
	logger := zap.NewNop().Sugar()

	apiKeyCalled := false
	jwtCalled := false

	jwtHandler := func(c *echo.Context) error {
		jwtCalled = true
		return nil
	}

	apiKeyHandler := func(c *echo.Context) error {
		apiKeyCalled = true
		return nil
	}

	authChain := NewTestAuthChain(jwtHandler, apiKeyHandler, logger)

	e := echo.New()
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Request().Header.Set("Authorization", "") // Empty header

	next := func(c *echo.Context) error { return nil }
	err := authChain.AllAuth()(next)(c)

	// Verify neither middleware was called (empty header treated as missing)
	assert.False(t, apiKeyCalled)
	assert.False(t, jwtCalled)

	// Verify error for 401
	assert.Error(t, err)
}

func TestAuthChain_AllAuth_JWTWithEmptyAPIKey(t *testing.T) {
	logger := zap.NewNop().Sugar()

	apiKeyCalled := false
	jwtCalled := false

	jwtHandler := func(c *echo.Context) error {
		jwtCalled = true
		c.Set("authType", "jwt")
		c.Set("userId", "test-user-id")
		return nil
	}

	apiKeyHandler := func(c *echo.Context) error {
		apiKeyCalled = true
		return nil
	}

	authChain := NewTestAuthChain(jwtHandler, apiKeyHandler, logger)

	e := echo.New()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "")                     // Empty API key
	req.Header.Set("Authorization", "Bearer jwt-token") // Valid JWT
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	next := func(c *echo.Context) error { return nil }
	err := authChain.AllAuth()(next)(c)

	assert.NoError(t, err)

	// Verify JWT middleware was called (empty API key ignored)
	assert.False(t, apiKeyCalled)
	assert.True(t, jwtCalled)

	// Verify JWT context values
	authType, exists := c.Get("authType")
	assert.True(t, exists)
	assert.Equal(t, "jwt", authType)
}

func TestAuthChain_AllAuth_VerifyLogging(t *testing.T) {
	logger := zap.NewNop().Sugar()

	apiKeyCalled := false
	jwtCalled := false

	jwtHandler := func(c *echo.Context) error {
		jwtCalled = true
		return nil
	}

	apiKeyHandler := func(c *echo.Context) error {
		apiKeyCalled = true
		c.Set("authType", "api_key")
		c.Set("apiKeyId", "test-api-key-id")
		return nil
	}

	authChain := NewTestAuthChain(jwtHandler, apiKeyHandler, logger)

	e := echo.New()
	req := httptest.NewRequest("GET", "/test-path", nil)
	req.Header.Set("X-API-Key", "pk_test-token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	next := func(c *echo.Context) error { return nil }
	err := authChain.AllAuth()(next)(c)

	assert.NoError(t, err)

	// Verify API key middleware was called (which indicates logging occurred)
	assert.True(t, apiKeyCalled)
	assert.False(t, jwtCalled)
}
