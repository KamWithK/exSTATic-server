package database

import (
	"database/sql"

	_ "github.com/libsql/libsql-client-go/libsql"
	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB(dbUrl string) error {
	var err error

	DB, err = sql.Open("libsql", dbUrl)
	if err != nil {
		return err
	}

	return DB.Ping()
}
