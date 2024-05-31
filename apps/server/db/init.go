package db

import (
	"downite/cmd/migrations"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

var DB *sqlx.DB

func DbInit() error {
	var err error
	DB, err = sqlx.Connect("sqlite", filepath.Join(".", "bin", "downite.db"))
	if err != nil {
		panic(err)
	}

	err = DB.Ping()
	if err != nil {
		panic(err)
	}
	migrationsDir := filepath.Join(".", "db", "migrations")

	err = migrations.Migrate(DB, migrationsDir)
	if err != nil {
		panic(err)
	}

	return nil
}
