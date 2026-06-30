package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type Account struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	Email        string     `json:"email"`
	DisplayName  string     `json:"display_name"`
	PhotoURL     string     `json:"photo_url"`
	Provider     string     `json:"provider"`
	IMAPHost     string     `json:"imap_host"`
	IMAPPort     int        `json:"imap_port"`
	SMTPHost     string     `json:"smtp_host"`
	SMTPPort     int        `json:"smtp_port"`
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token,omitempty"`
	TokenExpiry  *time.Time `json:"token_expiry,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type Session struct {
	ID        string    `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Folder struct {
	Name      string   `json:"name"`
	Delimiter string   `json:"delimiter"`
	Flags     []string `json:"flags"`
}

type Message struct {
	UID            int64     `json:"uid"`
	Subject        string    `json:"subject"`
	From           string    `json:"from"`
	Date           time.Time `json:"date"`
	Snippet        string    `json:"snippet"`
	IsRead         bool      `json:"is_read"`
	IsStarred      bool      `json:"is_starred"`
	HasAttachments bool      `json:"has_attachments"`
	Size           int       `json:"size"`
}
