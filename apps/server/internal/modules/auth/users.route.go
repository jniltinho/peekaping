package auth

import (
	"github.com/labstack/echo/v5"
)

type UsersRoute struct {
	controller *Controller
	middleware *MiddlewareProvider
}

func NewUsersRoute(
	controller *Controller,
	middleware *MiddlewareProvider,
) *UsersRoute {
	return &UsersRoute{controller, middleware}
}

func (r *UsersRoute) ConnectRoute(rg *echo.Group, controller *Controller) {
	users := rg.Group("/users")
	users.Use(r.middleware.Auth())
	users.GET("", controller.ListUsers)
	users.POST("", controller.CreateUser)
	users.DELETE("/:id", controller.DeleteUser)
	users.PATCH("/:id/active", controller.SetUserActive)
	users.PUT("/:id/password", controller.ResetUserPassword)

	regSettings := rg.Group("/settings")
	regSettings.Use(r.middleware.Auth())
	regSettings.GET("/registration", controller.GetRegistrationStatus)
	regSettings.PUT("/registration", controller.UpdateRegistrationStatus)
}
