package setting

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
		controller, middleware,
	}
}

func (uc *Route) ConnectRoute(
	rg *echo.Group,
	controller *Controller,
) {
	router := rg.Group("/settings")

	router.Use(uc.middleware.AllAuth())

	router.GET("/key/:key", uc.controller.GetByKey)
	router.PUT("/key/:key", uc.controller.SetByKey)
	router.DELETE("/key/:key", uc.controller.DeleteByKey)
}
