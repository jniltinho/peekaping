package middleware

import (
	"net/http"
	"peekaping/internal/modules/api_key"
	"peekaping/internal/modules/auth"
	"peekaping/internal/utils"

	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

// MARK: Types and Constructor

// AuthChain provides a chained authentication middleware
type AuthChain struct {
	jwtMiddleware    *auth.MiddlewareProvider
	apiKeyMiddleware *api_key.MiddlewareProvider
	logger           *zap.SugaredLogger
}

// NewAuthChain creates a new authentication chain
func NewAuthChain(
	jwtMiddleware *auth.MiddlewareProvider,
	apiKeyMiddleware *api_key.MiddlewareProvider,
	logger *zap.SugaredLogger,
) *AuthChain {
	return &AuthChain{
		jwtMiddleware:    jwtMiddleware,
		apiKeyMiddleware: apiKeyMiddleware,
		logger:           logger.Named("[auth_chain]"),
	}
}

// MARK: AllAuth

// AllAuth creates a middleware that supports both JWT and API key authentication
// The middleware automatically routes requests based on header presence:
// - If X-API-Key header is present: routes to API key authentication
// - Otherwise: routes to JWT authentication (expects Authorization header with Bearer token)
func (ac *AuthChain) AllAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			apiKeyHeader := c.Request().Header.Get("X-API-Key")
			authHeader := c.Request().Header.Get("Authorization")

			if apiKeyHeader != "" {
				// Route to API key authentication
				ac.logger.Debugw("Routing to API key authentication", "ip", c.RealIP(), "path", c.Request().URL.Path, "keyPrefix", apiKeyHeader[:min(len(apiKeyHeader), 10)]+"...")
				// Call the inner middleware (which returns a MiddlewareFunc)
				return ac.apiKeyMiddleware.Auth()(next)(c)
			} else if authHeader != "" {
				// Route to JWT authentication
				ac.logger.Debugw("Routing to JWT authentication", "ip", c.RealIP(), "path", c.Request().URL.Path, "tokenPrefix", authHeader[:min(len(authHeader), 10)]+"...")
				return ac.jwtMiddleware.Auth()(next)(c)
			} else {
				// No authentication headers provided
				ac.logger.Warnw("Missing authentication headers", "ip", c.RealIP(), "path", c.Request().URL.Path)
				return c.JSON(http.StatusUnauthorized, utils.NewFailResponse("Authentication required: provide either X-API-Key header or Authorization header"))
			}
		}
	}
}
