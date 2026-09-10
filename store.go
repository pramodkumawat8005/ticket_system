package main

import (
	"sync"
	"time"
)


type Store struct {
	mu          sync.RWMutex
	usersByID   map[string]*User
	usersByName map[string]*User
	tickets     map[string]*Ticket
}

func NewStore() *Store {
	return &Store{
		usersByID:   make(map[string]*User),
		usersByName: make(map[string]*User),
		tickets:     make(map[string]*Ticket),
	}
}


func (s *Store) CreateUser(username, passwordHash string) (*User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.usersByName[username]; exists {
		return nil, false
	}

	u := &User{
		ID:           newID(),
		Username:     username,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now().UTC(),
	}
	s.usersByID[u.ID] = u
	s.usersByName[u.Username] = u
	return u, true
}


func (s *Store) GetUserByUsername(username string) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.usersByName[username]
	return u, ok
}


func (s *Store) CreateTicket(userID, title, description string) *Ticket {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	
	t := &Ticket{
		ID:          newID(),
		UserID:      userID,
		Title:       title,
		Description: description,
		Status:      StatusOpen,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.tickets[t.ID] = t
	return t
}


func (s *Store) ListTicketsForUser(userID string) []*Ticket {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Ticket, 0)
	for _, t := range s.tickets {
		if t.UserID == userID {
			result = append(result, t)
		}
	}
	return result
}


func (s *Store) GetTicket(id string) (*Ticket, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tickets[id]
	return t, ok
}


func (s *Store) UpdateTicketStatus(id string, status TicketStatus) (*Ticket, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, ok := s.tickets[id]
	if !ok {
		return nil, false
	}
	t.Status = status
	t.UpdatedAt = time.Now().UTC()
	return t, true
}
