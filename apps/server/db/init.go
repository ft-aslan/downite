package db

import (
	"downite/cmd/migrations"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

var DB *sqlx.DB

func DbInit() (*sqlx.DB, error) {
	var err error
	db, err := sqlx.Connect("sqlite", filepath.Join(".", "bin", "downite.db"))
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	migrationsDir := filepath.Join(".", "db", "migrations")

	err = migrations.Migrate(db, migrationsDir)
	if err != nil {
		panic(err)
	}
	DB = db
	return db, nil
}
