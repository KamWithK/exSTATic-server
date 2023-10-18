package main

import (
	"log/slog"
	"net/http"

	_ "github.com/libsql/libsql-client-go/libsql"
	_ "modernc.org/sqlite"

	"github.com/KamWithK/exSTATic-backend/internal/database"
	"github.com/KamWithK/exSTATic-backend/internal/settings"
)

func main() {
	// var dbUrl = "http://127.0.0.1:8080"
	var dbUrl = "file:/data/local.db"
	err := database.InitDB(dbUrl)
	if err != nil {
		slog.Error("database couldn't connect", err)
	}
	defer database.DB.Close()

	http.HandleFunc("/", settings.SettingsHandler)

	httpErr := http.ListenAndServe(":8080", nil)
	if httpErr != nil {
		slog.Error("http server error", httpErr)
	}
}
