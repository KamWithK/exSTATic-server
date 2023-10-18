package settings

import "net/http"
import (
	"database/sql"
	_ "github.com/libsql/libsql-client-go/libsql"
	_ "modernc.org/sqlite"
)

func SettingsHandler(w http.ResponseWriter, r *http.Request) {
	// var dbUrl = "http://127.0.0.1:8080"
	var dbUrl = "file:path/to/file.db"
	db, err := sql.Open("libsql", dbUrl)
	if err != nil {

	}
	defer db.Close()

	w.Write([]byte("test"))
}
