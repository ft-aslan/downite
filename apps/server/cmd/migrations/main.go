package migrations

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
)

// Migrate runs all SQL migration files found in the migrations directory.
func Migrate(db *sqlx.DB, migrationsDir string) error {
	fmt.Printf("Migrating %s\n", migrationsDir)
	//TODO(fatih): we need to embed migrations with go:embed
	if err := goose.SetDialect("sqlite"); err != nil {
		panic(err)
	}
	err := goose.Up(db.DB, migrationsDir)
	if err != nil {
		return err
	}
	return nil
}
