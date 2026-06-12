package notification_sent_history

import (
	"peekaping/internal/utils"

	"github.com/uptrace/bun"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

func RegisterDependencies(container *dig.Container) {
	// Register repository (MongoDB backend removed; always SQL)
	utils.RegisterRepository(
		container,
		func(db *bun.DB) Repository {
			return NewSQLRepository(db)
		},
	)

	// Register service
	container.Provide(func(
		repository Repository,
		logger *zap.SugaredLogger,
	) Service {
		return NewService(repository, logger)
	})
}
