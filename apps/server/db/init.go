package db

import (
	"downite/cmd/migrations"
	"downite/utils"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type Database struct {
	x *sqlx.DB
}

func DbInit() (*Database, error) {
	var err error
	projectRoot, err := utils.FindProjectRoot()
	if err != nil {
		return nil, err
	}
	x, err := sqlx.Connect("sqlite", filepath.Join(projectRoot, "bin", "downite.db"))
	if err != nil {
		panic(err)
	}

	err = x.Ping()
	if err != nil {
		panic(err)
	}
	migrationsDir := filepath.Join(projectRoot, "db", "migrations")

	err = migrations.Migrate(x, migrationsDir)
	if err != nil {
		panic(err)
	}
	db := &Database{
		x,
	}
	return db, nil
}
