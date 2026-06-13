package migrations

import (
	"embed"

	"github.com/uptrace/bun/migrate"
)

var Migrations = migrate.NewMigrations()

//go:embed *.sql
var sqlFiles embed.FS

func init() {
	if err := Migrations.Discover(sqlFiles); err != nil {
		panic(err)
	}
}
