package db

import (
	"downite/cmd/migrations"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type Database struct {
	x *sqlx.DB
}

func DbInit() (*Database, error) {
	var err error
	x, err := sqlx.Connect("sqlite", filepath.Join(".", "bin", "downite.db"))
	if err != nil {
		panic(err)
	}

	err = x.Ping()
	if err != nil {
		panic(err)
	}
	migrationsDir := filepath.Join(".", "db", "migrations")

	err = migrations.Migrate(x, migrationsDir)
	if err != nil {
		panic(err)
	}
	db := &Database{
		x,
	}
	return db, nil
}
