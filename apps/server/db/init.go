package db

import (
	_ "database/sql"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var x *sqlx.DB

func DbInit() error {
	var err error
	x, err = sqlx.Connect("sqlite3", "./downite.db")
	if err != nil {
		panic(err)
	}

	err = x.Ping()
	if err != nil {
		panic(err)
	}

	return nil
}
