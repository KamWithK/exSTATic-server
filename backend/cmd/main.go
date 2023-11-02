package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	_ "github.com/libsql/libsql-client-go/libsql"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
	_ "modernc.org/sqlite"

	"github.com/KamWithK/exSTATic-backend/internal/auth"
	"github.com/KamWithK/exSTATic-backend/internal/database"
	"github.com/KamWithK/exSTATic-backend/internal/settings"
)

var GoogleOAuthConfig = oauth2.Config{
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	RedirectURL:  os.Getenv("DOMAIN_URL") + "/callback",
	Scopes:       []string{"email", "profile", "openid"},
	Endpoint:     endpoints.Google,
}

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

	// Auth configuration
	provider, err := oidc.NewProvider(context.Background(), "https://accounts.google.com")
	oidcConfig := &oidc.Config{
		ClientID: os.Getenv("GOOGLE_CLIENT_ID"),
	}
	verifier := provider.Verifier(oidcConfig)
	googleOIDCConfig := auth.Auth{
		OAuthConfig: GoogleOAuthConfig,
		Provider:    provider,
		Verifier:    *verifier,
	}

	// Login
	mux.HandleFunc("/login", googleOIDCConfig.LoginHandler)
	mux.HandleFunc("/callback", googleOIDCConfig.CallbackHandler)

	// APIs
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/settings", settings.SettingsHandler)
	mux.Handle("/api/", http.StripPrefix("/api", googleOIDCConfig.AuthMiddleware(apiMux)))

	// Serve
	httpErr := http.ListenAndServe(":8080", corsMiddleware(mux))
	if httpErr != nil {
		slog.Error("http server error", httpErr)
	}
}
