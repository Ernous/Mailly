package api

import (
	"net/http"

	"github.com/ernela/mailly/internal/store"
)

type AuthHandler struct {
	storage store.Storage
}

func NewAuthHandler(storage store.Storage) *AuthHandler {
	return &AuthHandler{storage: storage}
}

func (h *AuthHandler) Check(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("mailly_session")
	if err != nil {
		jsonOK(w, map[string]interface{}{"authenticated": false})
		return
	}

	sess, err := h.storage.GetSession(r.Context(), cookie.Value)
	if err != nil {
		jsonOK(w, map[string]interface{}{"authenticated": false})
		return
	}

	jsonOK(w, map[string]interface{}{
		"authenticated": true,
		"user_id":       sess.UserID.String(),
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("mailly_session")
	if err == nil {
		h.storage.DeleteSession(r.Context(), cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "mailly_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	jsonOK(w, map[string]string{"status": "ok"})
}
