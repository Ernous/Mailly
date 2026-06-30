package store

import (
	"context"
	"time"

	"github.com/ernela/mailly/internal/models"
	"github.com/google/uuid"
)

type Storage interface {
	// Users
	GetOrCreateUser(ctx context.Context, email string) (*models.User, error)

	// Accounts
	CreateAccount(ctx context.Context, a *models.Account) error
	GetAccountsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Account, error)
	GetAccountByID(ctx context.Context, id uuid.UUID) (*models.Account, error)
	UpdateAccountToken(ctx context.Context, id uuid.UUID, accessToken string, tokenExpiry *time.Time) error
	DeleteAccount(ctx context.Context, id uuid.UUID) error

	// Sessions
	CreateSession(ctx context.Context, userID uuid.UUID) (string, error)
	GetSession(ctx context.Context, sessionID string) (*models.Session, error)
	DeleteSession(ctx context.Context, sessionID string) error
}
