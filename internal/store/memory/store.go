package memory

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ernela/mailly/internal/models"
	"github.com/google/uuid"
)

type Store struct {
	mu       sync.RWMutex
	users    map[string]*models.User
	accounts map[uuid.UUID]*models.Account
	sessions map[string]*models.Session
}

func New() *Store {
	return &Store{
		users:    make(map[string]*models.User),
		accounts: make(map[uuid.UUID]*models.Account),
		sessions: make(map[string]*models.Session),
	}
}

func (s *Store) GetOrCreateUser(ctx context.Context, email string) (*models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if u, ok := s.users[email]; ok {
		return u, nil
	}

	u := &models.User{
		ID:        uuid.New(),
		Email:     email,
		CreatedAt: time.Now(),
	}
	s.users[email] = u
	return u, nil
}

func (s *Store) CreateAccount(ctx context.Context, a *models.Account) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Deduplicate: update tokens if account with same email+provider already exists
	for _, existing := range s.accounts {
		if existing.Email == a.Email && existing.Provider == a.Provider {
			log.Printf("[STORE] Account already exists: %s (%s), updating tokens", a.Email, a.Provider)
			existing.AccessToken = a.AccessToken
			existing.RefreshToken = a.RefreshToken
			existing.TokenExpiry = a.TokenExpiry
			*a = *existing
			return nil
		}
	}

	s.accounts[a.ID] = a
	return nil
}

func (s *Store) GetAccountsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Account, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []models.Account
	for _, a := range s.accounts {
		if a.UserID == userID {
			result = append(result, *a)
		}
	}
	return result, nil
}

func (s *Store) GetAccountByID(ctx context.Context, id uuid.UUID) (*models.Account, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	a, ok := s.accounts[id]
	if !ok {
		return nil, fmt.Errorf("account not found")
	}
	return a, nil
}

func (s *Store) UpdateAccountToken(ctx context.Context, id uuid.UUID, accessToken string, tokenExpiry *time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	a, ok := s.accounts[id]
	if !ok {
		return fmt.Errorf("account not found")
	}
	a.AccessToken = accessToken
	a.TokenExpiry = tokenExpiry
	return nil
}

func (s *Store) DeleteAccount(ctx context.Context, id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.accounts, id)
	return nil
}

func (s *Store) CreateSession(ctx context.Context, userID uuid.UUID) (string, error) {
	b := make([]byte, 32)
	rand.Read(b)
	id := hex.EncodeToString(b)

	s.mu.Lock()
	s.sessions[id] = &models.Session{
		ID:        id,
		UserID:    userID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	s.mu.Unlock()
	return id, nil
}

func (s *Store) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sess, ok := s.sessions[sessionID]
	if !ok || time.Now().After(sess.ExpiresAt) {
		return nil, fmt.Errorf("session not found or expired")
	}
	return sess, nil
}

func (s *Store) DeleteSession(ctx context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionID)
	return nil
}
