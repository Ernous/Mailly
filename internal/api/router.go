package api

import (
	"net/http"

	"github.com/ernela/mailly/internal/oauth"
	"github.com/ernela/mailly/internal/store"
)

func NewRouter(storage store.Storage, stateStore *oauth.StateStore) http.Handler {
	authHandler := NewAuthHandler(storage)
	accountHandler := NewAccountHandler(storage, stateStore)
	messageHandler := NewMessageHandler(storage)

	mux := http.NewServeMux()

	// Auth
	mux.HandleFunc("GET /api/auth/check", authHandler.Check)
	mux.HandleFunc("POST /api/auth/logout", authHandler.Logout)

	// OAuth
	mux.HandleFunc("GET /api/providers", accountHandler.Providers)
	mux.HandleFunc("POST /api/accounts/connect", accountHandler.Connect)
	mux.HandleFunc("POST /api/accounts/custom", accountHandler.ConnectCustom)
	mux.HandleFunc("GET /api/accounts/callback", accountHandler.Callback)
	mux.HandleFunc("GET /api/accounts", accountHandler.List)
	mux.HandleFunc("DELETE /api/accounts/{id}", accountHandler.Delete)

	// IMAP proxy
	mux.HandleFunc("GET /api/folders", messageHandler.ListFolders)
	mux.HandleFunc("GET /api/messages", messageHandler.ListByFolder)
	mux.HandleFunc("GET /api/message", messageHandler.GetByID)
	mux.HandleFunc("POST /api/message/mark-read", messageHandler.MarkRead)
	mux.HandleFunc("POST /api/message/mark-unread", messageHandler.MarkUnread)
	mux.HandleFunc("DELETE /api/message", messageHandler.DeleteMessage)

	// Quota
	mux.HandleFunc("GET /api/quota", accountHandler.Quota)

	var handler http.Handler = mux
	handler = SessionMiddleware(storage)(handler)
	handler = corsMiddleware(handler)

	return handler
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Mailly-*")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
