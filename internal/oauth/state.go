package oauth

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type StateStore struct {
	states map[string]*StateEntry
}

type StateEntry struct {
	State     string
	Provider  Provider
	CreatedAt time.Time
}

func NewStateStore() *StateStore {
	s := &StateStore{states: make(map[string]*StateEntry)}
	go s.cleanup()
	return s
}

func (s *StateStore) Create(provider Provider) string {
	state := uuid.New().String()
	s.states[state] = &StateEntry{
		State:     state,
		Provider:  provider,
		CreatedAt: time.Now(),
	}
	return state
}

func (s *StateStore) Get(state string) (*StateEntry, error) {
	entry, ok := s.states[state]
	if !ok {
		return nil, fmt.Errorf("invalid or expired state")
	}
	if time.Since(entry.CreatedAt) > 5*time.Minute {
		delete(s.states, state)
		return nil, fmt.Errorf("state expired")
	}
	delete(s.states, state)
	return entry, nil
}

func (s *StateStore) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		for k, v := range s.states {
			if now.Sub(v.CreatedAt) > 5*time.Minute {
				delete(s.states, k)
			}
		}
	}
}
