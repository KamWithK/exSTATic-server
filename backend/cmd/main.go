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

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		next.ServeHTTP(w, r)
	})
}

func EmptyHandler(w http.ResponseWriter, r *http.Request) {
}

func main() {
	err := database.InitDB(os.Getenv("DB_URL"))
	if err != nil {
		slog.Error("database couldn't connect", err)
	}
	defer database.DB.Close()

	mux := http.NewServeMux()

	// Static content
	fs := http.FileServer(http.Dir("../static"))
	mux.Handle("/", fs)

	// APIs
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/settings", settings.SettingsHandler)
	apiMux.HandleFunc("/login", auth.LoginHandler)
	apiMux.HandleFunc("/callback", auth.CallbackHandler)
	mux.Handle("/api/", http.StripPrefix("/api", apiMux))

	// Serve
	httpErr := http.ListenAndServe(":8080", corsMiddleware(mux))
	if httpErr != nil {
		slog.Error("http server error", httpErr)
	}
}
