package api_key

import (
	"peekaping/internal/config"
	"peekaping/internal/utils"

	"go.uber.org/dig"
)

func RegisterDependencies(container *dig.Container, cfg *config.Config) {
	utils.RegisterRepository(container, NewSQLRepository)

	container.Provide(NewRoute)
	container.Provide(NewService)
	container.Provide(NewController)
	container.Provide(NewMiddlewareProvider)
}
