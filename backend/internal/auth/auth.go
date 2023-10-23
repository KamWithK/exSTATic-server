package auth

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/KamWithK/exSTATic-backend/internal/database"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
)

var GoogleOAuthConfig = oauth2.Config{
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	RedirectURL:  "http://localhost:8080/callback",
	Scopes:       []string{"email", "profile", "openid"},
	Endpoint:     endpoints.Google,
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Get and use verifier challenge
	// verifier := oauth2.GenerateVerifier()
	// url := GoogleOAuthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))
	url := GoogleOAuthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	// Get Google token
	state, code, verifier := r.FormValue("state"), r.FormValue("code"), r.FormValue("verifier")
	_ = verifier // TODO: Remove once verifier works

	if state != "state" {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		slog.Warn("invalid state")
		return
	}

	token, err := GoogleOAuthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		slog.Warn("could not exchange auth code", err)
		return
	}

	// Get client to make requests with
	client := GoogleOAuthConfig.Client(r.Context(), token)

	// User info request
	resp, err := client.Get("https://www.googleapis.com/oauth2/v1/userinfo")
	if err != nil {
		slog.Warn("could not retrieve user info")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	defer resp.Body.Close()

	// Read user info
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Warn("could not read user info")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Parse user info
	var userInfo map[string]interface{}
	err = json.Unmarshal(body, &userInfo)
	if err != nil {
		slog.Warn("could not parse user info")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	email, ok := userInfo["email"].(string)
	if !ok {
		slog.Warn("could not get the users email")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	if email == "" {
		slog.Warn("no registered email")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	name, ok := userInfo["name"].(string)

	// Check whether user registered in database
	registered := false
	err = database.DB.QueryRow("SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)", email).Scan(&registered)
	if err != nil {
		slog.Warn("database error during select user email exists: ", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Add user if not pre-existing
	if !registered {
		_, err := database.DB.Exec("INSERT INTO users (email, name) VALUES ($1, $2)", email, name)
		if err != nil {
			slog.Warn("database error during insert user: ", err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
	}
}
