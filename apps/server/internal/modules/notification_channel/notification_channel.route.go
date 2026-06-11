package notification_channel

import (
	"peekaping/internal/modules/middleware"

	"github.com/labstack/echo/v5"
)

type Route struct {
	controller *Controller
	middleware *middleware.AuthChain
}

func NewRoute(
	controller *Controller,
	middleware *middleware.AuthChain,
) *Route {
	return &Route{
		controller,
		middleware,
	}
}

func (uc *Route) ConnectRoute(
	rg *echo.Group,
	controller *Controller,
) {
	router := rg.Group("notification-channels")

	router.Use(uc.middleware.AllAuth())

	router.GET("", controller.FindAll)
	router.POST("", controller.Create)
	router.POST("/test", controller.Test)
	router.GET("/:id", controller.FindByID)
	router.PUT("/:id", controller.UpdateFull)
	router.PATCH("/:id", controller.UpdatePartial)
	router.DELETE("/:id", controller.Delete)
}
