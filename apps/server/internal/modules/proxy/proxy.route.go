package proxy

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
	router := rg.Group("proxies")

	router.Use(uc.middleware.AllAuth())
	router.GET("", uc.controller.FindAll)
	router.POST("", uc.controller.Create)
	router.GET(":id", uc.controller.FindByID)
	router.PUT(":id", uc.controller.UpdateFull)
	router.PATCH(":id", uc.controller.UpdatePartial)
	router.DELETE(":id", uc.controller.Delete)
}
