package auth

import (
	// "encoding/base64"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/KamWithK/exSTATic-backend/internal/database"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
)

var GoogleOAuthConfig = oauth2.Config{
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	RedirectURL:  os.Getenv("DOMAIN_URL") + "/api/callback",
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

	// Get tokens
	access_token, err := GoogleOAuthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		slog.Warn("could not exchange auth code", err)
		return
	}
	refresh_token := access_token.Extra("refresh_token").(string)
	id_token := strings.Split(access_token.Extra("id_token").(string), ".")

	if len(id_token) != 3 {
		slog.Error("incorrect id token length, the id token should be comprised off a header, body and signature")
		return
	}
	id_token_parts := make(map[int]map[string]interface{})

	for index, part := range id_token[:2] {
		part_json := make(map[string]interface{})
		part_bytes, err := base64.RawStdEncoding.DecodeString(part)
		if err != nil {
			slog.Warn("could not base64 decode JWT", err)
			return
		}
		err = json.Unmarshal(part_bytes, &part_json)
		if err != nil {
			slog.Warn("could not unmarshal JWT", err)
			// return
		}

		id_token_parts[index] = part_json
	}

	id_token_header := id_token_parts[0]
	id_token_body := id_token_parts[1]

	fmt.Println(id_token_header)
	fmt.Println(id_token_body)
	println(refresh_token)

	// Parse user info
	email := id_token_body["email"]
	if email == "" {
		slog.Warn("no registered email")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	name := id_token_body["name"]

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

	// Redirect back to the homepage
	// TODO: Generate and send an api auth token
	http.Redirect(w, r, os.Getenv("DOMAIN_URL"), http.StatusTemporaryRedirect)
}
