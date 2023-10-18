package main

import (
	"log"
	"net/http"

	"github.com/KamWithK/exSTATic-backend/internal/settings"
)

func main() {
	http.HandleFunc("/", settings.SettingsHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
