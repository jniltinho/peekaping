package auth

import (
	"net/http"
	"peekaping/internal/utils"
	"strings"

	"github.com/labstack/echo/v5"
)

// MiddlewareProvider holds all middleware functions
type MiddlewareProvider struct {
	tokenMaker *TokenMaker
}

// NewMiddlewareProvider creates a new middleware provider
func NewMiddlewareProvider(tokenMaker *TokenMaker) *MiddlewareProvider {
	return &MiddlewareProvider{
		tokenMaker: tokenMaker,
	}
}

// Auth is a middleware that verifies JWT access tokens only
func (p *MiddlewareProvider) Auth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			// Get the Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, utils.NewFailResponse("Authorization header is required"))
			}

			// JWT authentication
			// Add Bearer prefix if not present
			if !strings.HasPrefix(authHeader, "Bearer ") {
				authHeader = "Bearer " + authHeader
			}

			// Check if the header has the Bearer prefix
			fields := strings.Fields(authHeader)
			if len(fields) != 2 || fields[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, utils.NewFailResponse("Invalid authorization header format"))
			}

			// Extract the token
			accessToken := fields[1]

			// Verify the token
			claims, err := p.tokenMaker.VerifyToken(c.Request().Context(), accessToken, "access")
			if err != nil {
				return c.JSON(http.StatusUnauthorized, utils.NewFailResponse("Invalid or expired token"))
			}

			// Check if it's an access token
			if claims.Type != "access" {
				return c.JSON(http.StatusUnauthorized, utils.NewFailResponse("Invalid token type"))
			}

			// Set user information in the context
			c.Set("userId", claims.UserID)
			c.Set("email", claims.Email)
			c.Set("authType", "jwt")

			return next(c)
		}
	}
}
