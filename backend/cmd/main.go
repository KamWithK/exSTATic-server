package main

import (
	"log/slog"
	"net/http"
	"os"

	_ "github.com/libsql/libsql-client-go/libsql"
	_ "modernc.org/sqlite"

	"github.com/KamWithK/exSTATic-backend/internal/auth"
	"github.com/KamWithK/exSTATic-backend/internal/database"
	"github.com/KamWithK/exSTATic-backend/internal/settings"
)

func main() {
	err := database.InitDB(os.Getenv("DB_URL"))
	if err != nil {
		slog.Error("database couldn't connect", err)
	}
	defer database.DB.Close()

	http.HandleFunc("/", settings.SettingsHandler)
	http.HandleFunc("/login", auth.LoginHandler)
	http.HandleFunc("/callback", auth.CallbackHandler)

	httpErr := http.ListenAndServe(":8080", nil)
	if httpErr != nil {
		slog.Error("http server error", httpErr)
	}
}
