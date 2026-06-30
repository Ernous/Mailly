package api

import (
	"context"
	"log"
	"net/http"

	"github.com/ernela/mailly/internal/store"
	"github.com/google/uuid"
)

type contextKey string

const userIDKey contextKey = "userID"

func UserIDFromContext(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(userIDKey).(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}

func SessionMiddleware(storage store.Storage) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			// Public routes — no session required
			if path == "/api/auth/check" || path == "/api/auth/logout" ||
				path == "/api/providers" || r.Method == "OPTIONS" ||
				path == "/api/accounts/connect" || path == "/api/accounts/callback" {
				next.ServeHTTP(w, r)
				return
			}

			// Connect requires session
			cookie, err := r.Cookie("mailly_session")
			if err != nil {
				log.Printf("[AUTH] %s %s - no session cookie", r.Method, path)
				jsonError(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			sess, err := storage.GetSession(r.Context(), cookie.Value)
			if err != nil {
				log.Printf("[AUTH] %s %s - invalid session: %v", r.Method, path, err)
				jsonError(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			log.Printf("[AUTH] %s %s - user=%s", r.Method, path, sess.UserID)
			ctx := context.WithValue(r.Context(), userIDKey, sess.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
