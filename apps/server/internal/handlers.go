package internal

import (
	"net/http"
	"peekaping/version"

	"github.com/labstack/echo/v5"
)

// @Summary      Get server version
// @Description  Returns the current server version
// @Tags         System
// @Produce      json
// @Success      200  {object}  map[string]string  "{"version": "1.2.3"}"
// @Router       /version [get]
func versionHandler(c *echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"version": version.Version})
}

// @Summary      Get server health
// @Description  Returns the current server health
// @Tags         System
// @Produce      json
// @Success      200  {object}  map[string]string  "{"status": "success"}"
// @Router       /health [get]
func healthHandler(c *echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "success"})
}

func envJSHandler(c *echo.Context) error {
	js := "/* generated at startup */\nwindow.__CONFIG__ = {\n  API_URL: \"\"\n};\n"
	c.Response().Header().Set("Content-Type", "application/javascript")
	c.Response().Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	return c.String(http.StatusOK, js)
}
