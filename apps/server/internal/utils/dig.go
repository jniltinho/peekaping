package utils

import (
	"go.uber.org/dig"
)

// RegisterRepository registers the SQL repository implementation.
// MongoDB backend support has been removed.
func RegisterRepository(container *dig.Container, sqlRepoConstructor interface{}) {
	container.Provide(sqlRepoConstructor)
}
