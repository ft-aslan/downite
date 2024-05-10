package migrations

import (
	"downite/utils"
	"log"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
)

// Migrate runs all SQL migration files found in the migrations directory.
func Migrate(db *sqlx.DB, migrationsDir string) error {
	rows, err := db.Queryx("SELECT name FROM migrations ORDER BY id")
	if err != nil {
		return err
	}
	defer rows.Close()

	var appliedMigrations []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		appliedMigrations = append(appliedMigrations, name)
	}

	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return err
	}

	for _, file := range files {
		if utils.Contains(appliedMigrations, filepath.Base(file)) {
			continue
		}

		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		tx := db.MustBegin()
		_, err = tx.Exec(string(sqlBytes))
		if err != nil {
			tx.Rollback()
			return err
		}

		err = tx.Commit()
		if err != nil {
			return err
		}

		_, fileName := filepath.Split(file)
		_, err = db.Exec("INSERT INTO migrations (name) VALUES ($1)", fileName)
		if err != nil {
			return err
		}

		log.Printf("Applied migration: %s", fileName)
	}

	return nil
}
