package tag

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

func (r *Route) ConnectRoute(
	rg *echo.Group,
	controller *Controller,
) {
	router := rg.Group("/tags")

	router.Use(r.middleware.AllAuth())

	router.GET("", controller.FindAll)
	router.POST("", controller.Create)
	router.GET("/:id", controller.FindByID)
	router.PUT("/:id", controller.UpdateFull)
	router.PATCH("/:id", controller.UpdatePartial)
	router.DELETE("/:id", controller.Delete)
}
