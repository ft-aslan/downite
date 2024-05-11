package db

import (
	"downite/cmd/migrations"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

var x *sqlx.DB

func DbInit() error {
	var err error
	x, err = sqlx.Connect("sqlite", "./tmp/downite.db")
	if err != nil {
		panic(err)
	}

	err = x.Ping()
	if err != nil {
		panic(err)
	}
	migrationsDir := filepath.Join(".", "db", "migrations")

	sqlx.MustExec(x, "CREATE TABLE IF NOT EXISTS migrations (id SERIAL PRIMARY KEY, name TEXT UNIQUE)")

	err = migrations.Migrate(x, migrationsDir)
	if err != nil {
		panic(err)
	}

	return nil
}
