package api_key

import (
	"net/http"
	"peekaping/internal/utils"
	"strings"

	"github.com/labstack/echo/v5"
)

// MiddlewareProvider holds API key middleware functions
type MiddlewareProvider struct {
	service Service
}

// NewMiddlewareProvider creates a new API key middleware provider
func NewMiddlewareProvider(service Service) *MiddlewareProvider {
	return &MiddlewareProvider{
		service: service,
	}
}

// Auth is a middleware that verifies API key authentication
// This should be used as the final middleware in a chain for API key-only endpoints
func (p *MiddlewareProvider) Auth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			// Get the X-API-Key header
			authHeader := c.Request().Header.Get("X-API-Key")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, utils.NewFailResponse("X-API-Key header is required"))
			}

			// Only accept API keys
			if !strings.HasPrefix(authHeader, ApiKeyPrefix) {
				return c.JSON(http.StatusUnauthorized, utils.NewFailResponse("API key required"))
			}

			// Validate the API key
			apiKey, err := p.service.ValidateKey(c.Request().Context(), authHeader)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, utils.NewFailResponse("Invalid or expired API key"))
			}

			// Set API key information in the context
			c.Set("apiKeyId", apiKey.ID)
			c.Set("authType", "api_key")

			return next(c)
		}
	}
}
