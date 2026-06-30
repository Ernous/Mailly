package badger

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/ernela/mailly/internal/models"
	"github.com/google/uuid"
)

type Store struct {
	db *badger.DB
}

func New(path string) (*Store, error) {
	opts := badger.DefaultOptions(path)
	opts.Logger = nil
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("badger open: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func userKey(email string) []byte          { return []byte("user:" + email) }
func accountKey(id uuid.UUID) []byte       { return []byte("account:" + id.String()) }
func userAccountsKey(uid uuid.UUID) []byte { return []byte("user_accounts:" + uid.String()) }
func sessionKey(id string) []byte          { return []byte("session:" + id) }

func (s *Store) GetOrCreateUser(ctx context.Context, email string) (*models.User, error) {
	key := userKey(email)
	var user models.User

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			return err
		}
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &user)
		})
	})

	if err == badger.ErrKeyNotFound {
		user = models.User{
			ID:        uuid.New(),
			Email:     email,
			CreatedAt: time.Now(),
		}
		val, _ := json.Marshal(user)
		err = s.db.Update(func(txn *badger.Txn) error {
			return txn.Set(key, val)
		})
		return &user, err
	}

	return &user, err
}

func (s *Store) CreateAccount(ctx context.Context, a *models.Account) error {
	return s.db.Update(func(txn *badger.Txn) error {
		uaKey := userAccountsKey(a.UserID)
		aKey := accountKey(a.ID)
		var accountIDs []string

		item, err := txn.Get(uaKey)
		if err == nil {
			item.Value(func(val []byte) error {
				return json.Unmarshal(val, &accountIDs)
			})
		}

		for _, idStr := range accountIDs {
			aid, err := uuid.Parse(idStr)
			if err != nil {
				continue
			}
			item, err := txn.Get(accountKey(aid))
			if err != nil {
				continue
			}
			var existing models.Account
			item.Value(func(val []byte) error {
				return json.Unmarshal(val, &existing)
			})
			if existing.Email == a.Email && existing.Provider == a.Provider {
				log.Printf("[STORE] Account already exists: %s (%s), updating tokens", a.Email, a.Provider)
				existing.AccessToken = a.AccessToken
				existing.RefreshToken = a.RefreshToken
				existing.TokenExpiry = a.TokenExpiry
				val, _ := json.Marshal(existing)
				if err := txn.Set(aKey, val); err != nil {
					return err
				}
				*a = existing
				return nil
			}
		}

		val, _ := json.Marshal(a)
		if err := txn.Set(aKey, val); err != nil {
			return err
		}

		accountIDs = append(accountIDs, a.ID.String())
		idxVal, _ := json.Marshal(accountIDs)
		return txn.Set(uaKey, idxVal)
	})
}

func (s *Store) GetAccountsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Account, error) {
	var accounts []models.Account

	err := s.db.View(func(txn *badger.Txn) error {
		uaKey := userAccountsKey(userID)
		item, err := txn.Get(uaKey)
		if err == badger.ErrKeyNotFound {
			return nil
		}
		if err != nil {
			return err
		}

		var accountIDs []string
		item.Value(func(val []byte) error {
			return json.Unmarshal(val, &accountIDs)
		})

		for _, idStr := range accountIDs {
			aid, err := uuid.Parse(idStr)
			if err != nil {
				continue
			}
			item, err := txn.Get(accountKey(aid))
			if err != nil {
				continue
			}
			var acc models.Account
			item.Value(func(val []byte) error {
				return json.Unmarshal(val, &acc)
			})
			accounts = append(accounts, acc)
		}
		return nil
	})

	return accounts, err
}

func (s *Store) GetAccountByID(ctx context.Context, id uuid.UUID) (*models.Account, error) {
	var account models.Account

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(accountKey(id))
		if err == badger.ErrKeyNotFound {
			return fmt.Errorf("account not found")
		}
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &account)
		})
	})

	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (s *Store) UpdateAccountToken(ctx context.Context, id uuid.UUID, accessToken string, tokenExpiry *time.Time) error {
	return s.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get(accountKey(id))
		if err != nil {
			return fmt.Errorf("account not found")
		}
		var account models.Account
		item.Value(func(val []byte) error {
			return json.Unmarshal(val, &account)
		})
		account.AccessToken = accessToken
		account.TokenExpiry = tokenExpiry
		val, _ := json.Marshal(account)
		return txn.Set(accountKey(id), val)
	})
}

func (s *Store) DeleteAccount(ctx context.Context, id uuid.UUID) error {
	return s.db.Update(func(txn *badger.Txn) error {
		aKey := accountKey(id)
		item, err := txn.Get(aKey)
		if err != nil {
			return nil
		}

		var acc models.Account
		item.Value(func(val []byte) error {
			return json.Unmarshal(val, &acc)
		})

		if err := txn.Delete(aKey); err != nil {
			return err
		}

		uaKey := userAccountsKey(acc.UserID)
		var accountIDs []string
		item, err = txn.Get(uaKey)
		if err == nil {
			item.Value(func(val []byte) error {
				return json.Unmarshal(val, &accountIDs)
			})
		}

		for i, sid := range accountIDs {
			if sid == id.String() {
				accountIDs = append(accountIDs[:i], accountIDs[i+1:]...)
				break
			}
		}
		idxVal, _ := json.Marshal(accountIDs)
		return txn.Set(uaKey, idxVal)
	})
}

func (s *Store) CreateSession(ctx context.Context, userID uuid.UUID) (string, error) {
	b := make([]byte, 32)
	rand.Read(b)
	id := hex.EncodeToString(b)

	sess := &models.Session{
		ID:        id,
		UserID:    userID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	val, _ := json.Marshal(sess)
	err := s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(sessionKey(id), val)
	})

	if err != nil {
		return "", err
	}
	return id, nil
}

func (s *Store) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	var sess models.Session

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(sessionKey(sessionID))
		if err == badger.ErrKeyNotFound {
			return fmt.Errorf("session not found or expired")
		}
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &sess)
		})
	})

	if err != nil {
		return nil, err
	}

	if time.Now().After(sess.ExpiresAt) {
		s.db.Update(func(txn *badger.Txn) error {
			return txn.Delete(sessionKey(sessionID))
		})
		return nil, fmt.Errorf("session not found or expired")
	}

	return &sess, nil
}

func (s *Store) DeleteSession(ctx context.Context, sessionID string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(sessionKey(sessionID))
	})
}
