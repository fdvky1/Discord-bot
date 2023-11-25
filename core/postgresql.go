package core

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/fdvky1/Discord-bot/entity"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

var postgresDB *bun.DB

func NewPostgresDB() *bun.DB {
	if postgresDB != nil {
		return postgresDB
	}
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(os.Getenv("POSTGRESQL_URL"))))
	postgresDB = bun.NewDB(sqldb, pgdialect.New())
	postgresDB.SetMaxOpenConns(1)

	MigrateTables(&entity.DisabledCmdEntity{})
	return postgresDB
}

func MigrateTables(values ...interface{}) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	for _, value := range values {
		if _, err := postgresDB.NewCreateTable().Model(value).IfNotExists().Exec(ctx); err != nil {
			log.Fatal(err)
		}
	}
}
