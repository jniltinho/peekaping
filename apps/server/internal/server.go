package internal

import (
	"net/http"
	_ "peekaping/docs"
	"peekaping/internal/config"
	"peekaping/internal/modules/api_key"
	"peekaping/internal/modules/auth"
	"peekaping/internal/modules/badge"
	"peekaping/internal/modules/healthcheck"
	"peekaping/internal/modules/heartbeat"
	"peekaping/internal/modules/maintenance"
	"peekaping/internal/modules/monitor"
	"peekaping/internal/modules/notification_channel"
	"peekaping/internal/modules/proxy"
	"peekaping/internal/modules/queue"
	"peekaping/internal/modules/setting"
	"peekaping/internal/modules/status_page"
	"peekaping/internal/modules/tag"
	"peekaping/internal/modules/websocket"
	"peekaping/internal/version"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

// @Summary      Get server version
// @Description  Returns the current server version
// @Tags         System
// @Produce      json
// @Success      200  {object}  map[string]string  "{"version": "1.2.3"}"
// @Router       /version [get]
func versionHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"version": version.Version})
}

// @Summary      Get server health
// @Description  Returns the current server health
// @Tags         System
// @Produce      json
// @Success      200  {object}  map[string]string  "{"status": "success"}"
// @Router       /health [get]
func healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "success"})
}

type Server struct {
	Router *echo.Echo // Now an Echo instance (field name kept for minimal main.go changes)
	Cfg    *config.Config
}

func ProvideServer(
	logger *zap.SugaredLogger,
	cfg *config.Config,
	monitorRoute *monitor.MonitorRoute,
	monitorController *monitor.MonitorController,
	authRoute *auth.Route,
	authController *auth.Controller,
	wsServer *websocket.Server,
	notificationChannelRoute *notification_channel.Route,
	notificationChannelController *notification_channel.Controller,
	proxyRoute *proxy.Route,
	proxyController *proxy.Controller,
	settingRoute *setting.Route,
	settingController *setting.Controller,
	heartbeatService heartbeat.Service,
	monitorService monitor.Service,
	queueService queue.Service,
	maintenanceRoute *maintenance.Route,
	maintenanceController *maintenance.Controller,
	statusPageRoute *status_page.Route,
	statusPageController *status_page.Controller,
	tagRoute *tag.Route,
	tagController *tag.Controller,
	badgeRoute *badge.Route,
	badgeController *badge.Controller,
	apiKeyRoute *api_key.Route,
	apiKeyController *api_key.Controller,
) *Server {
	// Initialize Echo server
	e := echo.New()

	// In production-ish modes, we still want recovery.
	// Echo's Recover middleware is excellent.
	e.Use(middleware.Recover())

	// Optional logger in dev (Echo's built-in)
	if cfg.Mode == "dev" {
		e.Use(middleware.Logger())
	}

	// CORS - using Echo's built-in middleware (replaces gin-contrib/cors)
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions, http.MethodPatch},
		AllowHeaders:     []string{"Origin", "X-Requested-With", "Content-Type", "Accept", "Authorization", "X-API-Key"},
		ExposeHeaders:    []string{"Authorization"},
		AllowCredentials: true,
	}))

	// Simple health and version at root and under /api/v1
	e.GET("/health", healthHandler)
	e.GET("/version", versionHandler)

	api := e.Group("/api/v1")
	api.GET("/health", healthHandler)
	api.GET("/version", versionHandler)

	// Connect all module routes (signatures will be updated to *echo.Group in their route files)
	monitorRoute.ConnectRoute(api, monitorController)
	authRoute.ConnectRoute(api, authController)
	notificationChannelRoute.ConnectRoute(api, notificationChannelController)
	proxyRoute.ConnectRoute(api, proxyController)
	settingRoute.ConnectRoute(api, settingController)
	maintenanceRoute.ConnectRoute(api, maintenanceController)
	statusPageRoute.ConnectRoute(api, statusPageController)
	tagRoute.ConnectRoute(api, tagController)
	badgeRoute.ConnectRoute(api, badgeController)
	apiKeyRoute.ConnectRoute(api, apiKeyController)

	// Register push endpoint (will be updated to accept *echo.Group)
	healthcheck.RegisterPushEndpoint(api, monitorService, heartbeatService, queueService, logger)

	// Swagger
	// We removed gin-swagger. For now serve a basic handler.
	// Full interactive Swagger UI can be added later by mounting static assets
	// or adding a small echo-swagger dependency. The generated doc.json is still produced by swag.
	e.GET("/swagger/*", func(c echo.Context) error {
		path := c.Param("*")
		if path == "doc.json" || path == "doc.json/" {
			// The actual spec is generated into docs/swagger.json by swag.
			// For a working port we return a helpful message; teams usually copy
			// the json or serve the static UI separately.
			return c.JSON(http.StatusOK, map[string]any{
				"message": "Swagger spec available after 'swag init'. Mount /swagger/doc.json or use static UI assets.",
			})
		}
		return c.String(http.StatusOK, "Swagger UI placeholder. Replace with static assets or echo-swagger handler as needed.")
	})

	// WebSocket / socket.io compatibility shims (raw writer/request delegation)
	// Works the same way with Echo's access to underlying http objects.
	e.Any("/socket.io/*f", func(c echo.Context) error {
		wsServer.ServeHTTP(c.Response().Writer, c.Request())
		return nil
	})

	return &Server{Router: e, Cfg: cfg}
}
