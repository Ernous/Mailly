package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func jsonOK(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

type Credentials struct {
	Provider    string
	Email       string
	IMAPHost    string
	IMAPPort    int
	AccessToken string
}

func parseCredentials(r *http.Request) (*Credentials, error) {
	provider := r.Header.Get("X-Mailly-Provider")
	host := r.Header.Get("X-Mailly-IMAP-Host")
	portStr := r.Header.Get("X-Mailly-IMAP-Port")
	token := r.Header.Get("X-Mailly-Access-Token")
	email := r.Header.Get("X-Mailly-Email")

	if host == "" || token == "" {
		return nil, fmt.Errorf("missing mailly credentials")
	}

	port, _ := strconv.Atoi(portStr)
	if port == 0 {
		port = 993
	}

	return &Credentials{
		Provider:    provider,
		Email:       email,
		IMAPHost:    host,
		IMAPPort:    port,
		AccessToken: token,
	}, nil
}
