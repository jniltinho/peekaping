package monitor

import (
	"peekaping/internal/config"
	"peekaping/internal/utils"

	"go.uber.org/dig"
)

func RegisterDependencies(container *dig.Container, cfg *config.Config) {
	utils.RegisterRepository(container, NewSQLRepository)
	container.Provide(NewMonitorService)
	container.Provide(NewMonitorController)
	container.Provide(NewMonitorRoute)
	container.Provide(NewMonitorEventListener)
}
