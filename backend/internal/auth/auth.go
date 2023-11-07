package auth

import (
	"context"
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
	// Create random state and nonce for this device
	state, err := randString(16)
	if err != nil {
		slog.Error("could not create random string for OAuth2 state")
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	nonce, err := randString(16)
	if err != nil {
		slog.Error("could not create random string for OIDC nonce")
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Put state and nonce into cookies so they can be checked within the callback
	stateCookie := &http.Cookie{
		Name:     "state",
		Value:    state,
		MaxAge:   int(time.Hour.Seconds()),
		Secure:   r.TLS != nil,
		HttpOnly: true,
	}
	nonceCookie := &http.Cookie{
		Name:     "nonce",
		Value:    nonce,
		MaxAge:   int(time.Hour.Seconds()),
		Secure:   r.TLS != nil,
		HttpOnly: true,
	}

	http.SetCookie(w, stateCookie)
	http.SetCookie(w, nonceCookie)

	// Initiate redirect to start login
	http.Redirect(w, r, auth.OAuthConfig.AuthCodeURL(state, oidc.Nonce(nonce), oauth2.AccessTypeOffline), http.StatusTemporaryRedirect)
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

	// Get the nonce
	nonce, err := r.Cookie("nonce")
	if err != nil {
		slog.Warn("nonce not found")
	}

	// Attempt token exchange
	tokens, err := auth.OAuthConfig.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		slog.Warn("could not exchange auth code", err)
		return
	}

	// Extract and verify the nonce
	rawIDToken, ok := tokens.Extra("id_token").(string)
	if !ok {
		slog.Warn("failed to parse ID token", err)
		return
	}
	idToken, err := auth.Verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		slog.Warn("id token verification failed")
		return
	}
	if idToken.Nonce != nonce.Value {
		slog.Warn("invalid OIDC nonce")
		return
	}

	// Get refresh token
	refreshToken, ok := tokens.Extra("refresh_token").(string)
	if !ok {
		slog.Warn("failed to parse refresh token", err)
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

	// Send id token and refresh token back as cookies
	refreshCookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		MaxAge:   int(time.Hour.Seconds()),
		Secure:   r.TLS != nil,
		HttpOnly: true,
	}
	idCookie := &http.Cookie{
		Name:     "id_token",
		Value:    rawIDToken,
		MaxAge:   int(time.Hour.Seconds()),
		Secure:   r.TLS != nil,
		HttpOnly: true,
	}

	// Set the cookies
	http.SetCookie(w, refreshCookie)
	http.SetCookie(w, idCookie)
}

func (auth *Auth) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Context for future use
		ctx := r.Context()

		// Ensure users first need to login
		authenticated := false
		defer func() {
			if !authenticated {
				http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			} else {
				next.ServeHTTP(w, r.WithContext(ctx))
			}
		}()

		// Get tokens from cookies
		rawIDToken, err := r.Cookie("id_token")
		if err != nil {
			slog.Warn("id token not found", err)
			slog.Info("will continue in case refresh token works instead")
			// return
		}
		refreshToken, err := r.Cookie("refresh_token")
		if err != nil {
			slog.Warn("refresh token not found")
			return
		}

		// Reconstruct old token
		oldToken := &oauth2.Token{
			RefreshToken: refreshToken.Value,
		}
		if rawIDToken != nil {
			oldToken = oldToken.WithExtra(map[string]string{"id_token": rawIDToken.Value})
		}

		// Refresh token when needed
		tokenSource := auth.OAuthConfig.TokenSource(ctx, oldToken)
		token, err := tokenSource.Token()
		if err != nil {
			slog.Warn("failed to refresh token", err)
			return
		}

		// Verify id token
		idToken, err := auth.Verifier.Verify(ctx, token.Extra("id_token").(string))
		if err != nil {
			slog.Warn("id token verification failed", err)
			return
		}

		// Get user claims and ensure they're valid
		var claims struct {
			Email         string `json:"email"`
			EmailVerified bool   `json:"email_verified"`
		}
		if err := idToken.Claims(&claims); err != nil {
			slog.Warn("claims from OIDC id token are invalid", err)
			return
		}
		if claims.Email == "" {
			slog.Warn("no registered email")
			return
		}
		if !claims.EmailVerified {
			slog.Warn("email not verified")
			return
		}

		// Make sure user is name
		name := ""
		user_id := -1
		err = database.DB.QueryRow("SELECT rowid, name FROM users WHERE email = $1", claims.Email).Scan(&user_id, &name)
		if err != nil {
			slog.Warn("database error during select user", err)
			return
		}
		if user_id < 0 {
			slog.Error("invalid user id or user id not found: ", user_id)
			return
		}

		// Add user email into context for future use
		ctx = context.WithValue(ctx, "email", claims.Email)
		ctx = context.WithValue(ctx, "name", name)

		// Set the user as authenticated
		authenticated = true
	})
}
