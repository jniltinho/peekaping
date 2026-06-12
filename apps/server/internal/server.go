package internal

import (
	"net/http"
	_ "peekaping/docs"
	"peekaping/internal/config"
	"peekaping/internal/frontend"
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

	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"go.uber.org/zap"
)

type Server struct {
	Router *echo.Echo
	Cfg    *config.Config
}

func ProvideServer(
	logger *zap.SugaredLogger,
	cfg *config.Config,
	monitorRoute *monitor.MonitorRoute,
	monitorController *monitor.MonitorController,
	authRoute *auth.Route,
	authUsersRoute *auth.UsersRoute,
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
	e := echo.New()

	e.Use(middleware.Recover())
	if cfg.Mode == "dev" {
		e.Use(middleware.RequestLogger())
	}

	// AllowCredentials must NOT be true when AllowOrigins contains "*".
	// The app uses Authorization: Bearer (no cookies), so "*" is safe.
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions, http.MethodPatch},
		AllowHeaders:  []string{"Origin", "X-Requested-With", "Content-Type", "Accept", "Authorization", "X-API-Key"},
		ExposeHeaders: []string{"Authorization"},
	}))

	e.GET("/health", healthHandler)
	e.GET("/version", versionHandler)

	api := e.Group("/api/v1")
	api.GET("/health", healthHandler)
	api.GET("/version", versionHandler)

	monitorRoute.ConnectRoute(api, monitorController)
	authRoute.ConnectRoute(api, authController)
	authUsersRoute.ConnectRoute(api, authController)
	notificationChannelRoute.ConnectRoute(api, notificationChannelController)
	proxyRoute.ConnectRoute(api, proxyController)
	settingRoute.ConnectRoute(api, settingController)
	maintenanceRoute.ConnectRoute(api, maintenanceController)
	statusPageRoute.ConnectRoute(api, statusPageController)
	tagRoute.ConnectRoute(api, tagController)
	badgeRoute.ConnectRoute(api, badgeController)
	apiKeyRoute.ConnectRoute(api, apiKeyController)

	healthcheck.RegisterPushEndpoint(api, monitorService, heartbeatService, queueService, logger)

	e.GET("/swagger/*", echo.WrapHandler(httpSwagger.WrapHandler))

	e.Any("/socket.io/*f", func(c *echo.Context) error {
		wsServer.ServeHTTP(c.Response(), c.Request())
		return nil
	})

	// env.js and SPA catch-all must come last
	e.GET("/env.js", envJSHandler)
	e.GET("/*", frontend.Handler())

	return &Server{Router: e, Cfg: cfg}
}
