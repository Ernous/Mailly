package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/emersion/go-imap"
	imapclient "github.com/emersion/go-imap/client"
	"github.com/emersion/go-imap/responses"
	"github.com/ernela/mailly/internal/models"
	"github.com/ernela/mailly/internal/oauth"
	"github.com/ernela/mailly/internal/store"
	"github.com/google/uuid"
)

type AccountHandler struct {
	storage    store.Storage
	stateStore *oauth.StateStore
}

func NewAccountHandler(storage store.Storage, stateStore *oauth.StateStore) *AccountHandler {
	return &AccountHandler{storage: storage, stateStore: stateStore}
}

func (h *AccountHandler) Providers(w http.ResponseWriter, r *http.Request) {
	log.Println("[API] GET /api/providers")
	var resp []map[string]string
	for name, cfg := range oauth.Providers {
		resp = append(resp, map[string]string{
			"name":         string(name),
			"display_name": cfg.DisplayName,
		})
	}
	if resp == nil {
		resp = []map[string]string{}
	}
	jsonOK(w, resp)
}

func (h *AccountHandler) Connect(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Provider string `json:"provider"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[API] POST /api/accounts/connect - invalid request: %v", err)
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}

	log.Printf("[API] POST /api/accounts/connect - provider: %s", req.Provider)

	provider := oauth.Provider(req.Provider)
	if _, ok := oauth.GetProviderConfig(provider); !ok {
		log.Printf("[API] POST /api/accounts/connect - unknown provider: %s", req.Provider)
		jsonError(w, "unknown provider", http.StatusBadRequest)
		return
	}

	state := h.stateStore.Create(provider)
	authURL, err := oauth.GetAuthURL(provider, state)
	if err != nil {
		log.Printf("[API] POST /api/accounts/connect - failed to generate auth URL: %v", err)
		jsonError(w, "failed to generate auth URL", http.StatusInternalServerError)
		return
	}

	log.Printf("[API] POST /api/accounts/connect - auth URL generated, state: %s", state)
	jsonOK(w, map[string]string{"redirect_url": authURL})
}

func (h *AccountHandler) Callback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	log.Printf("[API] GET /api/accounts/callback - state=%s, code=%s...", state, safeCode(code))

	if code == "" || state == "" {
		log.Printf("[API] GET /api/accounts/callback - missing code or state")
		jsonError(w, "missing code or state", http.StatusBadRequest)
		return
	}

	entry, err := h.stateStore.Get(state)
	if err != nil {
		log.Printf("[API] GET /api/accounts/callback - invalid state: %v", err)
		jsonError(w, "invalid state: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[API] GET /api/accounts/callback - exchanging code for provider: %s", entry.Provider)

	token, err := oauth.ExchangeCode(r.Context(), entry.Provider, code)
	if err != nil {
		log.Printf("[API] GET /api/accounts/callback - token exchange failed: %v", err)
		jsonError(w, "failed to exchange code: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[API] GET /api/accounts/callback - token received, getting user info...")

	userInfo, err := oauth.GetUserInfo(r.Context(), entry.Provider, token.AccessToken)
	if err != nil {
		log.Printf("[API] GET /api/accounts/callback - get user info failed: %v", err)
		jsonError(w, "failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[API] GET /api/accounts/callback - user info: email=%s, name=%s", userInfo.Email, userInfo.Name)

	var userID uuid.UUID
	var sessionID string

	// If there's an existing session, reuse its user (add account to same user)
	if cookie, err := r.Cookie("mailly_session"); err == nil {
		if sess, err := h.storage.GetSession(r.Context(), cookie.Value); err == nil {
			userID = sess.UserID
			sessionID = cookie.Value
			log.Printf("[API] GET /api/accounts/callback - reusing existing session: user=%s", userID)
		}
	}

	if userID == uuid.Nil {
		user, err := h.storage.GetOrCreateUser(r.Context(), userInfo.Email)
		if err != nil {
			log.Printf("[API] GET /api/accounts/callback - get or create user failed: %v", err)
			jsonError(w, "failed to create user", http.StatusInternalServerError)
			return
		}
		userID = user.ID

		sessionID, err = h.storage.CreateSession(r.Context(), userID)
		if err != nil {
			log.Printf("[API] GET /api/accounts/callback - create session failed: %v", err)
			jsonError(w, "failed to create session", http.StatusInternalServerError)
			return
		}
		log.Printf("[API] GET /api/accounts/callback - new session created: %s", sessionID[:8]+"...")
	}

	log.Printf("[API] GET /api/accounts/callback - user: id=%s", userID)

	// Save account with tokens
	providerCfg, _ := oauth.GetProviderConfig(entry.Provider)
	account := &models.Account{
		ID:           uuid.New(),
		UserID:       userID,
		Email:        userInfo.Email,
		DisplayName:  userInfo.Name,
		PhotoURL:     userInfo.PhotoURL,
		Provider:     string(entry.Provider),
		IMAPHost:     providerCfg.IMAPHost,
		IMAPPort:     providerCfg.IMAPPort,
		SMTPHost:     providerCfg.SMTPHost,
		SMTPPort:     providerCfg.SMTPPort,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}

	if !token.Expiry.IsZero() {
		exp := token.Expiry
		account.TokenExpiry = &exp
	}

	if err := h.storage.CreateAccount(r.Context(), account); err != nil {
		log.Printf("[API] GET /api/accounts/callback - create account failed: %v", err)
		jsonError(w, "failed to save account", http.StatusInternalServerError)
		return
	}

	log.Printf("[API] GET /api/accounts/callback - account created: id=%s, email=%s", account.ID, userInfo.Email)

	// Set session cookie (same domain as Go server = :3000)
	http.SetCookie(w, &http.Cookie{
		Name:     "mailly_session",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	})

	// Return HTML that posts message to opener and closes
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html><html><head><title>Mailly</title></head><body>
<script>
(function() {
  var accountId = %q;
  var email = %q;
  var provider = %q;
  function done() {
    try {
      if (window.opener && !window.opener.closed) {
        window.opener.postMessage(
          {type: "mailly:connected", account_id: accountId, email: email, provider: provider},
          window.location.origin
        );
      }
    } catch(e) {
      console.error("postMessage failed:", e);
    }
    setTimeout(function() { window.close(); }, 300);
  }
  // Small delay to ensure cookie is set
  setTimeout(done, 200);
})();
</script>
<h2 style="font-family:sans-serif;color:#4d8080">&#10003; Connected: %s</h2>
<p style="font-family:sans-serif;color:#999">This window will close automatically.</p>
</body></html>`, account.ID.String(), userInfo.Email, entry.Provider, userInfo.Email)
}

func (h *AccountHandler) ConnectCustom(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		IMAPHost string `json:"imap_host"`
		IMAPPort int    `json:"imap_port"`
		SMTPHost string `json:"smtp_host"`
		SMTPPort int    `json:"smtp_port"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" || req.IMAPHost == "" || req.IMAPPort == 0 {
		jsonError(w, "missing required fields", http.StatusBadRequest)
		return
	}

	// Test connection
	addr := fmt.Sprintf("%s:%d", req.IMAPHost, req.IMAPPort)
	c, err := imapclient.DialTLS(addr, nil)
	if err != nil {
		jsonError(w, "failed to connect to IMAP server: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := c.Login(req.Email, req.Password); err != nil {
		c.Close()
		jsonError(w, "invalid credentials: "+err.Error(), http.StatusUnauthorized)
		return
	}
	c.Logout()

	// Get or create user
	var userID uuid.UUID
	var sessionID string

	if cookie, err := r.Cookie("mailly_session"); err == nil {
		if sess, err := h.storage.GetSession(r.Context(), cookie.Value); err == nil {
			userID = sess.UserID
			sessionID = cookie.Value
		}
	}

	if userID == uuid.Nil {
		user, err := h.storage.GetOrCreateUser(r.Context(), req.Email)
		if err != nil {
			jsonError(w, "failed to create user", http.StatusInternalServerError)
			return
		}
		userID = user.ID

		sessionID, err = h.storage.CreateSession(r.Context(), userID)
		if err != nil {
			jsonError(w, "failed to create session", http.StatusInternalServerError)
			return
		}
	}

	account := &models.Account{
		ID:          uuid.New(),
		UserID:      userID,
		Email:       req.Email,
		DisplayName: req.Email,
		Provider:    "custom",
		IMAPHost:    req.IMAPHost,
		IMAPPort:    req.IMAPPort,
		SMTPHost:    req.SMTPHost,
		SMTPPort:    req.SMTPPort,
		AccessToken: req.Password, // Store password here for custom provider
	}

	if err := h.storage.CreateAccount(r.Context(), account); err != nil {
		jsonError(w, "failed to save account", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "mailly_session",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	})

	jsonOK(w, map[string]interface{}{"ok": true, "account_id": account.ID})
}

func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	log.Println("[API] GET /api/accounts")

	cookie, err := r.Cookie("mailly_session")
	if err != nil {
		log.Printf("[API] GET /api/accounts - no session cookie: %v", err)
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	sess, err := h.storage.GetSession(r.Context(), cookie.Value)
	if err != nil {
		log.Printf("[API] GET /api/accounts - invalid session: %v", err)
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	log.Printf("[API] GET /api/accounts - user_id=%s", sess.UserID)

	accounts, err := h.storage.GetAccountsByUserID(r.Context(), sess.UserID)
	if err != nil {
		log.Printf("[API] GET /api/accounts - failed: %v", err)
		jsonError(w, "failed to list accounts", http.StatusInternalServerError)
		return
	}

	var resp []map[string]interface{}
	for _, a := range accounts {
		resp = append(resp, map[string]interface{}{
			"id":           a.ID.String(),
			"email":        a.Email,
			"display_name": a.DisplayName,
			"photo_url":    a.PhotoURL,
			"provider":     a.Provider,
			"imap_host":    a.IMAPHost,
			"imap_port":    a.IMAPPort,
			"smtp_host":    a.SMTPHost,
			"smtp_port":    a.SMTPPort,
			"access_token": a.AccessToken,
		})
	}
	if resp == nil {
		resp = []map[string]interface{}{}
	}

	log.Printf("[API] GET /api/accounts - found %d accounts", len(resp))
	jsonOK(w, resp)
}

func (h *AccountHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}

	cookie, err := r.Cookie("mailly_session")
	if err != nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	sess, err := h.storage.GetSession(r.Context(), cookie.Value)
	if err != nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	account, err := h.storage.GetAccountByID(r.Context(), id)
	if err != nil || account.UserID != sess.UserID {
		jsonError(w, "account not found", http.StatusNotFound)
		return
	}

	log.Printf("[API] DELETE /api/accounts/%s", id)
	h.storage.DeleteAccount(r.Context(), id)
	jsonOK(w, map[string]string{"status": "deleted"})
}

func (h *AccountHandler) Quota(w http.ResponseWriter, r *http.Request) {
	creds, err := parseCredentials(r)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	creds, err = refreshCredentials(r.Context(), h.storage, creds)
	if err != nil {
		log.Printf("[IMAP] GET /api/quota - token refresh failed: %v", err)
		jsonError(w, err.Error(), http.StatusBadGateway)
		return
	}

	log.Printf("[IMAP] GET /api/quota - email=%s", creds.Email)

	c, err := connectIMAP(creds)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer c.Logout()

	caps, _ := c.Capability()
	if !caps["QUOTA"] {
		log.Printf("[IMAP] GET /api/quota - server does not support QUOTA extension")
		jsonOK(w, map[string]interface{}{"used": 0, "total": 0})
		return
	}

	var used, total uint64
	handler := responses.HandlerFunc(func(resp imap.Resp) error {
		data, ok := resp.(*imap.DataResp)
		if !ok || len(data.Fields) < 3 {
			return nil
		}
		name, _ := data.Fields[0].(string)
		if name != "QUOTA" {
			return nil
		}
		quotaList, ok := data.Fields[2].([]interface{})
		if !ok || len(quotaList) < 3 {
			return nil
		}
		resource, _ := quotaList[0].(string)
		if resource == "STORAGE" {
			if u, ok := quotaList[1].(uint32); ok {
				used = uint64(u)
			}
			if t, ok := quotaList[2].(uint32); ok {
				total = uint64(t)
			}
		}
		return nil
	})

	cmd := &imap.Command{Name: "GETQUOTAROOT", Arguments: []interface{}{"INBOX"}}
	_, err = c.Execute(cmd, handler)
	if err != nil {
		log.Printf("[IMAP] GET /api/quota - GETQUOTAROOT failed: %v", err)
		jsonOK(w, map[string]interface{}{"used": 0, "total": 0})
		return
	}

	jsonOK(w, map[string]interface{}{
		"used":  used,
		"total": total,
	})
}

func safeCode(code string) string {
	if len(code) > 20 {
		return code[:20] + "..."
	}
	return code
}
