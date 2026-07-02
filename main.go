package main

import (
	"log"
	"os"

	"signature-menu-backend/internal/app"
	"signature-menu-backend/internal/config"
	"signature-menu-backend/internal/patch"
	"signature-menu-backend/internal/store"
	"signature-menu-backend/pkg/token"
)

func main() {
	cfg := config.Load()

	dataStore, err := store.New(cfg.DataFile)
	if err != nil {
		log.Fatalf("failed to initialize data store: %v", err)
	}

	if len(os.Args) >= 3 && os.Args[1] == "patch" {
		if err := patch.Run(os.Args[2], dataStore); err != nil {
			log.Fatalf("patch %s failed: %v", os.Args[2], err)
		}
		return
	}

	tokens := token.NewManager(cfg.JWTSecret, "signature-menu", cfg.TokenTTL)
	router := app.NewRouter(cfg, dataStore, tokens)

	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
