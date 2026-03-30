package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"
)

type SessionStore struct {
	sessions map[string]*Session 
	mu       sync.RWMutex
}

type Session struct {
	Username  string
	ExpiresAt time.Time
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]*Session),
	}
}

func (s *SessionStore) CreateSession(username string) string {
	token := generateToken()
	
	s.mu.Lock()
	s.sessions[token] = &Session{
		Username:  username,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	s.mu.Unlock()
	
	return token
}

func (s *SessionStore) ValidateSession(token string) (string, bool) {
	s.mu.Lock()
	session, exists := s.sessions[token]
	
	
	if !exists {
		s.mu.Unlock()
		return "", false
	}

	if time.Now().After(session.ExpiresAt) {
		s.DeleteSession(token)
		s.mu.Unlock()
		return "", false
	}
	
	session.ExpiresAt = time.Now().Add(24 * time.Hour)

	s.mu.Unlock()
	return session.Username, true
}

func (s *SessionStore) DeleteSession(token string) {
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}

func generateToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (s *SessionStore) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Session-Token")
		if token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		
		username, valid := s.ValidateSession(token)
		if !valid {
			http.Error(w, "Session expired or invalid", http.StatusUnauthorized)
			return
		}
		
		ctx := context.WithValue(r.Context(), "username", username)
		next(w, r.WithContext(ctx))
	}
}