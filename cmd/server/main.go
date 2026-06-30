package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ernela/mailly/internal/api"
	"github.com/ernela/mailly/internal/config"
	"github.com/ernela/mailly/internal/oauth"
	"github.com/ernela/mailly/internal/store/badger"
)

func main() {
	cfg := config.Load()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	storage, err := badger.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to open BadgerDB at %s: %v", cfg.DBPath, err)
	}
	defer storage.Close()

	redirectURI := cfg.ServerURL + "/api/accounts/callback"
	oauth.RegisterGoogle(cfg.GoogleClientID, cfg.GoogleClientSecret, redirectURI)
	oauth.RegisterMicrosoft(cfg.MicrosoftClientID, cfg.MicrosoftClientSecret, redirectURI)

	stateStore := oauth.NewStateStore()
	router := api.NewRouter(storage, stateStore)

	server := &http.Server{
		Addr:         cfg.ServerAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Mailly server starting on %s", cfg.ServerAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Mailly stopped")
}
