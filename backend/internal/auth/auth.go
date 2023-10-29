package auth

import (
	// "encoding/base64"

	"crypto/rand"
	"encoding/base64"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/KamWithK/exSTATic-backend/internal/database"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type Auth struct {
	OAuthConfig oauth2.Config
	Provider    *oidc.Provider
	Verifier    oidc.IDTokenVerifier
}

// Source: https://github.com/coreos/go-oidc/blob/v3/example/idtoken/app.go
func randString(nByte int) (string, error) {
	b := make([]byte, nByte)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (auth *Auth) LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Create random state for this device
	state, err := randString(16)
	if err != nil {
		slog.Error("could not create random string for OAuth2 state")
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Put state into cookie so it can be checked within the callback
	cookie := &http.Cookie{
		Name:     "state",
		Value:    state,
		MaxAge:   int(time.Hour.Seconds()),
		Secure:   r.TLS != nil,
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)

	// Initiate redirect to start login
	http.Redirect(w, r, auth.OAuthConfig.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

func (auth *Auth) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	// End by redirecting to the main page
	defer http.Redirect(w, r, "/", http.StatusTemporaryRedirect)

	// Get and check state matches
	state, err := r.Cookie("state")
	if err != nil {
		slog.Warn("state not found")
		return
	}
	if r.URL.Query().Get("state") != state.Value {
		slog.Warn("invalid state")
		return
	}

	// Attempt token exchange
	tokens, err := auth.OAuthConfig.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		slog.Warn("could not exchange auth code", err)
		return
	}

	// Extract user info and claims
	userInfo, err := auth.Provider.UserInfo(r.Context(), oauth2.StaticTokenSource(tokens))
	if err != nil {
		slog.Warn("could not get user info", err)
		return
	}
	var claims struct {
		Name          string `json:"name"`
		EmailVerified bool   `json:"email_verified"`
	}
	userInfo.Claims(&claims)

	// Parse user info
	if userInfo.Email == "" {
		slog.Warn("no registered email")
		return
	}
	if !claims.EmailVerified {
		slog.Warn("email not verified")
		return
	}

	// Check whether user registered in database
	registered := false
	err = database.DB.QueryRow("SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)", userInfo.Email).Scan(&registered)
	if err != nil {
		slog.Warn("database error during select user email exists: ", err)
		return
	}

	// Add user if not pre-existing
	if !registered {
		_, err := database.DB.Exec("INSERT INTO users (email, name) VALUES ($1, $2)", userInfo.Email, claims.Name)
		if err != nil {
			slog.Warn("database error during insert user: ", err)
			return
		}
	}
}
