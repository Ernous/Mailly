package main

import (
	"context"
	"embed"
	"io/fs"
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

//go:embed dist
var staticFiles embed.FS

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
	apiRouter := api.NewRouter(storage, stateStore)

	// Serve the embedded frontend; fall back to index.html for SPA routing
	distFS, err := fs.Sub(staticFiles, "dist")
	if err != nil {
		log.Fatalf("Failed to create sub FS: %v", err)
	}
	fileServer := http.FileServer(http.FS(distFS))
	spaHandler := spaFileServer(fileServer, distFS)

	mux := http.NewServeMux()
	mux.Handle("/api/", apiRouter)
	mux.Handle("/", spaHandler)

	server := &http.Server{
		Addr:         cfg.ServerAddr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Mailly listening on %s", cfg.ServerAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
	log.Println("Mailly stopped")
}

// spaFileServer serves static files and falls back to index.html for unknown paths (SPA routing).
func spaFileServer(fileServer http.Handler, fsys fs.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" || path == "" {
			fileServer.ServeHTTP(w, r)
			return
		}
		// Check if file exists in the embedded FS
		f, err := fsys.Open(path[1:]) // strip leading /
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}
		// Not found — serve index.html for client-side routing
		r2 := r.Clone(r.Context())
		r2.URL.Path = "/"
		fileServer.ServeHTTP(w, r2)
	})
}
