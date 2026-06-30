package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
	"github.com/ernela/mailly/internal/models"
	"github.com/ernela/mailly/internal/oauth"
	"github.com/ernela/mailly/internal/store"
	"golang.org/x/oauth2"
)

type MessageHandler struct {
	storage store.Storage
}

func NewMessageHandler(storage store.Storage) *MessageHandler {
	return &MessageHandler{storage: storage}
}

func (h *MessageHandler) ListFolders(w http.ResponseWriter, r *http.Request) {
	creds, err := parseCredentials(r)
	if err != nil {
		log.Printf("[IMAP] GET /api/folders - bad credentials: %v", err)
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	creds, err = refreshCredentials(r.Context(), h.storage, creds)
	if err != nil {
		log.Printf("[IMAP] GET /api/folders - token refresh failed: %v", err)
		jsonError(w, err.Error(), http.StatusBadGateway)
		return
	}

	log.Printf("[IMAP] GET /api/folders - connecting to %s:%d as %s", creds.IMAPHost, creds.IMAPPort, creds.Email)
	c, err := connectIMAP(creds)
	if err != nil {
		log.Printf("[IMAP] GET /api/folders - connect failed: %v", err)
		jsonError(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer c.Logout()

	boxes := make(chan *imap.MailboxInfo, 50)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", boxes)
	}()

	var folders []map[string]interface{}
	for box := range boxes {
		name := box.Name
		if name == "FYI" {
			continue
		}
		displayName := strings.TrimPrefix(name, "[Gmail]/")
		displayName = strings.TrimPrefix(displayName, "[GoogleMail]/")
		folders = append(folders, map[string]interface{}{
			"name":      displayName,
			"full_name": name,
			"delimiter": box.Delimiter,
		})
	}
	<-done

	if folders == nil {
		folders = []map[string]interface{}{}
	}

	jsonOK(w, folders)
}

func (h *MessageHandler) ListByFolder(w http.ResponseWriter, r *http.Request) {
	creds, err := parseCredentials(r)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	creds, err = refreshCredentials(r.Context(), h.storage, creds)
	if err != nil {
		log.Printf("[IMAP] GET /api/messages - token refresh failed: %v", err)
		jsonError(w, err.Error(), http.StatusBadGateway)
		return
	}

	folderName := r.URL.Query().Get("folder")
	if folderName == "" {
		folderName = "INBOX"
	}

	log.Printf("[IMAP] GET /api/messages - folder=%s, email=%s", folderName, creds.Email)

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 50
	}

	c, err := connectIMAP(creds)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer c.Logout()

	status, err := c.Select(folderName, true)
	if err != nil {
		jsonError(w, "failed to select folder: "+err.Error(), http.StatusBadRequest)
		return
	}

	total := int(status.Messages)
	if total == 0 {
		jsonOK(w, map[string]interface{}{"messages": []interface{}{}, "total": 0})
		return
	}

	from := uint32(1)
	to := uint32(total)
	if limit > 0 && limit < total {
		from = uint32(total - limit + 1)
	}

	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	messages := make(chan *imap.Message, 100)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, imap.FetchRFC822Size, imap.FetchUid}, messages)
	}()

	var resp []map[string]interface{}
	for msg := range messages {
		if msg.Envelope == nil {
			continue
		}

		fromStr := ""
		if len(msg.Envelope.From) > 0 {
			a := msg.Envelope.From[0]
			fromStr = a.MailboxName + "@" + a.HostName
		}

		subj := msg.Envelope.Subject
		if len(subj) > 120 {
			subj = subj[:120] + "..."
		}

		isRead := false
		isStarred := false
		for _, f := range msg.Flags {
			if f == imap.SeenFlag {
				isRead = true
			}
			if f == imap.FlaggedFlag {
				isStarred = true
			}
		}

		resp = append(resp, map[string]interface{}{
			"uid":        msg.Uid,
			"subject":    subj,
			"from":       fromStr,
			"date":       msg.Envelope.Date,
			"is_read":    isRead,
			"is_starred": isStarred,
			"size":       msg.Size,
		})
	}
	<-done

	if resp == nil {
		resp = []map[string]interface{}{}
	}

	jsonOK(w, map[string]interface{}{
		"messages": resp,
		"total":    total,
	})
}

func (h *MessageHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	creds, err := parseCredentials(r)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	creds, err = refreshCredentials(r.Context(), h.storage, creds)
	if err != nil {
		log.Printf("[IMAP] GET /api/message - token refresh failed: %v", err)
		jsonError(w, err.Error(), http.StatusBadGateway)
		return
	}

	folderName := r.URL.Query().Get("folder")
	uidStr := r.URL.Query().Get("uid")
	if folderName == "" || uidStr == "" {
		jsonError(w, "folder and uid required", http.StatusBadRequest)
		return
	}

	uid, err := strconv.ParseUint(uidStr, 10, 32)
	if err != nil {
		jsonError(w, "invalid uid", http.StatusBadRequest)
		return
	}

	c, err := connectIMAP(creds)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer c.Logout()

	_, err = c.Select(folderName, true)
	if err != nil {
		jsonError(w, "failed to select folder", http.StatusBadRequest)
		return
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(uint32(uid))

	section := &imap.BodySectionName{}
	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, imap.FetchRFC822Size, section.FetchItem()}

	messages := make(chan *imap.Message, 1)
	done := make(chan error, 1)
	go func() {
		done <- c.UidFetch(seqset, items, messages)
	}()

	msg := <-messages
	<-done

	if msg == nil {
		jsonError(w, "message not found", http.StatusNotFound)
		return
	}

	fromStr := ""
	if len(msg.Envelope.From) > 0 {
		a := msg.Envelope.From[0]
		fromStr = a.MailboxName + "@" + a.HostName
	}

	toStr := ""
	for i, a := range msg.Envelope.To {
		if i > 0 {
			toStr += ", "
		}
		toStr += a.MailboxName + "@" + a.HostName
	}

	var textBody, htmlBody string
	cidImages := make(map[string]string)
	lit := msg.GetBody(section)
	if lit != nil {
		lr, err := mail.CreateReader(lit)
		if err == nil {
			for {
				p, err := lr.NextPart()
				if err != nil {
					break
				}
				b, err := io.ReadAll(p.Body)
				if err != nil {
					continue
				}
				contentType := p.Header.Get("Content-Type")
				contentID := p.Header.Get("Content-ID")
				if contentID != "" {
					cid := strings.Trim(contentID, "<>")
					mime := "application/octet-stream"
					if strings.Contains(contentType, "image/png") {
						mime = "image/png"
					} else if strings.Contains(contentType, "image/jpeg") || strings.Contains(contentType, "image/jpg") {
						mime = "image/jpeg"
					} else if strings.Contains(contentType, "image/gif") {
						mime = "image/gif"
					} else if strings.Contains(contentType, "image/svg") {
						mime = "image/svg+xml"
					} else if strings.Contains(contentType, "image/") {
						mime = contentType[:strings.Index(contentType, "image/") + len("image/")]
						if idx := strings.Index(mime, ";"); idx != -1 {
							mime = mime[:idx]
						}
					}
					cidImages[cid] = "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(b)
				} else if strings.Contains(contentType, "text/html") {
					htmlBody = string(b)
				} else if textBody == "" {
					textBody = string(b)
				}
			}
		}
	}

	if htmlBody != "" {
		for cid, dataURI := range cidImages {
			htmlBody = strings.ReplaceAll(htmlBody, "cid:"+cid, dataURI)
		}
		fromDomain := ""
		if idx := strings.Index(fromStr, "@"); idx != -1 {
			fromDomain = "https://" + fromStr[idx+1:]
		}
		if fromDomain != "" {
			htmlBody = strings.ReplaceAll(htmlBody, "src=\"/", "src=\""+fromDomain+"/")
			htmlBody = strings.ReplaceAll(htmlBody, "src=\"//", "src=\"https://")
			htmlBody = strings.ReplaceAll(htmlBody, "href=\"/ ", "href=\""+fromDomain+"/")
			htmlBody = strings.ReplaceAll(htmlBody, "background=\"/ ", "background=\""+fromDomain+"/")
		}
	}

	isRead := false
	isStarred := false
	for _, f := range msg.Flags {
		if f == imap.SeenFlag {
			isRead = true
		}
		if f == imap.FlaggedFlag {
			isStarred = true
		}
	}

	jsonOK(w, map[string]interface{}{
		"uid":        msg.Uid,
		"subject":    msg.Envelope.Subject,
		"from":       fromStr,
		"to":         toStr,
		"date":       msg.Envelope.Date,
		"text_body":  textBody,
		"html_body":  htmlBody,
		"is_read":    isRead,
		"is_starred": isStarred,
		"size":       msg.Size,
	})
}

func (h *MessageHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	creds, err := parseCredentials(r)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	creds, err = refreshCredentials(r.Context(), h.storage, creds)
	if err != nil {
		log.Printf("[IMAP] POST /api/message/mark-read - token refresh failed: %v", err)
		jsonError(w, err.Error(), http.StatusBadGateway)
		return
	}

	folderName := r.URL.Query().Get("folder")
	uidStr := r.URL.Query().Get("uid")
	if folderName == "" || uidStr == "" {
		jsonError(w, "folder and uid required", http.StatusBadRequest)
		return
	}

	uid, err := strconv.ParseUint(uidStr, 10, 32)
	if err != nil {
		jsonError(w, "invalid uid", http.StatusBadRequest)
		return
	}

	c, err := connectIMAP(creds)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer c.Logout()

	_, err = c.Select(folderName, false)
	if err != nil {
		jsonError(w, "failed to select folder: "+err.Error(), http.StatusBadRequest)
		return
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(uint32(uid))

	flags := []interface{}{imap.SeenFlag}
	err = c.UidStore(seqset, imap.FormatFlagsOp(imap.AddFlags, true), flags, nil)
	if err != nil {
		jsonError(w, "failed to mark as read: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[IMAP] POST /api/message/mark-read - uid=%d folder=%s marked as read", uid, folderName)
	jsonOK(w, map[string]interface{}{"ok": true})
}

func (h *MessageHandler) MarkUnread(w http.ResponseWriter, r *http.Request) {
	creds, err := parseCredentials(r)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	creds, err = refreshCredentials(r.Context(), h.storage, creds)
	if err != nil {
		log.Printf("[IMAP] POST /api/message/mark-unread - token refresh failed: %v", err)
		jsonError(w, err.Error(), http.StatusBadGateway)
		return
	}

	folderName := r.URL.Query().Get("folder")
	uidStr := r.URL.Query().Get("uid")
	if folderName == "" || uidStr == "" {
		jsonError(w, "folder and uid required", http.StatusBadRequest)
		return
	}

	uid, err := strconv.ParseUint(uidStr, 10, 32)
	if err != nil {
		jsonError(w, "invalid uid", http.StatusBadRequest)
		return
	}

	c, err := connectIMAP(creds)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer c.Logout()

	_, err = c.Select(folderName, false)
	if err != nil {
		jsonError(w, "failed to select folder: "+err.Error(), http.StatusBadRequest)
		return
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(uint32(uid))

	flags := []interface{}{imap.SeenFlag}
	err = c.UidStore(seqset, imap.FormatFlagsOp(imap.RemoveFlags, true), flags, nil)
	if err != nil {
		jsonError(w, "failed to mark as unread: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[IMAP] POST /api/message/mark-unread - uid=%d folder=%s marked as unread", uid, folderName)
	jsonOK(w, map[string]interface{}{"ok": true})
}

func (h *MessageHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	creds, err := parseCredentials(r)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	creds, err = refreshCredentials(r.Context(), h.storage, creds)
	if err != nil {
		log.Printf("[IMAP] DELETE /api/message - token refresh failed: %v", err)
		jsonError(w, err.Error(), http.StatusBadGateway)
		return
	}

	folderName := r.URL.Query().Get("folder")
	uidStr := r.URL.Query().Get("uid")
	if folderName == "" || uidStr == "" {
		jsonError(w, "folder and uid required", http.StatusBadRequest)
		return
	}

	uid, err := strconv.ParseUint(uidStr, 10, 32)
	if err != nil {
		jsonError(w, "invalid uid", http.StatusBadRequest)
		return
	}

	c, err := connectIMAP(creds)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer c.Logout()

	_, err = c.Select(folderName, false)
	if err != nil {
		jsonError(w, "failed to select folder: "+err.Error(), http.StatusBadRequest)
		return
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(uint32(uid))

	// Determine trash folder name by provider
	trashFolder := trashFolderName(creds.Provider)

	if trashFolder != "" && folderName != trashFolder {
		// Move to trash: COPY to trash, then delete original
		if err := c.UidCopy(seqset, trashFolder); err != nil {
			// Trash folder might have a different name — fall back to expunge
			log.Printf("[IMAP] copy to trash failed (%q): %v — falling back to expunge", trashFolder, err)
			trashFolder = ""
		}
	}

	// Mark original as deleted
	flags := []interface{}{imap.DeletedFlag}
	if err := c.UidStore(seqset, imap.FormatFlagsOp(imap.AddFlags, true), flags, nil); err != nil {
		jsonError(w, "failed to mark as deleted: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Expunge the original
	if err := c.Expunge(nil); err != nil {
		log.Printf("[IMAP] expunge warning: %v", err)
	}

	log.Printf("[IMAP] DELETE /api/message - uid=%d folder=%s trash=%s", uid, folderName, trashFolder)
	jsonOK(w, map[string]interface{}{"ok": true})
}

// trashFolderName returns the IMAP trash folder name for the given provider.
func trashFolderName(provider string) string {
	switch provider {
	case "google":
		return "[Gmail]/Trash"
	case "microsoft":
		return "Deleted Items"
	default:
		return "Trash"
	}
}

func refreshCredentials(ctx context.Context, storage store.Storage, creds *Credentials) (*Credentials, error) {
	if creds.Provider == "custom" {
		return creds, nil
	}

	accounts, err := storage.GetAccountsByUserID(ctx, UserIDFromContext(ctx))
	if err != nil {
		log.Printf("[REFRESH] GetAccountsByUserID: %v", err)
		return creds, nil
	}

	var account *models.Account
	for _, a := range accounts {
		if a.Email == creds.Email && a.Provider == creds.Provider {
			account = &a
			break
		}
	}
	if account == nil {
		return creds, nil
	}

	if account.TokenExpiry != nil && time.Now().Before(*account.TokenExpiry) {
		return creds, nil
	}

	if account.RefreshToken == "" {
		return nil, fmt.Errorf("token expired and no refresh token available — please re-add the account")
	}

	newToken, err := oauth.RefreshToken(ctx, oauth.Provider(creds.Provider), &oauth2.Token{
		AccessToken:  account.AccessToken,
		RefreshToken: account.RefreshToken,
		Expiry:       *account.TokenExpiry,
	})
	if err != nil {
		return nil, fmt.Errorf("refresh token: %w", err)
	}

	now := time.Now()
	if err := storage.UpdateAccountToken(ctx, account.ID, newToken.AccessToken, &now); err != nil {
		return nil, fmt.Errorf("update token: %w", err)
	}

	creds.AccessToken = newToken.AccessToken
	return creds, nil
}

func connectIMAP(creds *Credentials) (*client.Client, error) {
	addr := fmt.Sprintf("%s:%d", creds.IMAPHost, creds.IMAPPort)

	c, err := client.DialTLS(addr, nil)
	if err != nil {
		return nil, fmt.Errorf("imap connect: %w", err)
	}

	if creds.Provider == "custom" {
		if err := c.Login(creds.Email, creds.AccessToken); err != nil {
			c.Close()
			return nil, fmt.Errorf("imap login: %w", err)
		}
		return c, nil
	}

	mechanism := "XOAUTH2"
	if creds.Provider == "microsoft" {
		mechanism = "OAUTHBEARER"
	}

	saslClient := NewSASLClient(creds.Email, creds.AccessToken, mechanism)
	if err := c.Authenticate(saslClient); err != nil {
		c.Close()
		return nil, fmt.Errorf("imap auth: %w", err)
	}

	return c, nil
}
