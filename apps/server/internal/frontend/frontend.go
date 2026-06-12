package frontend

import (
	"embed"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v5"
)

//go:embed all:web
var webFS embed.FS

// Handler returns an Echo handler that serves the embedded SPA.
// Unknown paths fall back to index.html to support client-side routing.
// Returns 503 when the frontend was not embedded at build time.
func Handler() echo.HandlerFunc {
	sub, err := fs.Sub(webFS, "web")
	if err != nil {
		panic("frontend: failed to sub webFS: " + err.Error())
	}

	return func(c *echo.Context) error {
		urlPath := strings.TrimPrefix(c.Request().URL.Path, "/")
		if urlPath == "" {
			urlPath = "index.html"
		}

		target, err := resolveTarget(sub, urlPath)
		if err != nil {
			return c.String(http.StatusServiceUnavailable,
				"frontend not embedded; run 'make embed-web' to build and embed the frontend")
		}

		setCacheHeaders(c, target)
		http.ServeFileFS(c.Response(), c.Request(), sub, target)
		return nil
	}
}

// resolveTarget returns the file path to serve: the requested path if it exists
// as a regular file, or "index.html" for SPA fallback. Returns an error only
// when index.html itself is missing (frontend not embedded).
func resolveTarget(sub fs.FS, urlPath string) (string, error) {
	f, err := sub.Open(urlPath)
	if err == nil {
		info, _ := f.Stat()
		f.Close()
		if info != nil && !info.IsDir() {
			return urlPath, nil
		}
	}
	// File missing or is a directory → SPA fallback
	if _, err2 := sub.Open("index.html"); err2 != nil {
		return "", err2
	}
	return "index.html", nil
}

func setCacheHeaders(c *echo.Context, filePath string) {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".js", ".mjs", ".css", ".woff", ".woff2", ".svg", ".png", ".jpg", ".jpeg", ".gif", ".ico":
		c.Response().Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	default:
		c.Response().Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	}
}
