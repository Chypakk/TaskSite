package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"tasksite/internal/repository"
	"tasksite/internal/storage"
	"time"
)

type SessionStore struct {
	repo repository.SessionRepository
	// sessions map[string]*Session
	// mu       sync.RWMutex
}

type Session struct {
	Username  string
	ExpiresAt time.Time
}

func NewSessionStore(storage *storage.Storage) *SessionStore {
	return &SessionStore{
		repo: storage,
	}
}

func (s *SessionStore) CreateSession(ctx context.Context, username string) string {
	token := generateToken()

	err := s.repo.CreateSession(ctx, token, username, time.Now().Add(24 * time.Hour))
	if err != nil {
		log.Printf("Failed to create session for %s: %v", username, err)
		return ""
	}
	return token
}

func (s *SessionStore) ValidateSession(ctx context.Context, token string) (string, bool) {
	
	session, err := s.repo.GetSessionByToken(ctx, token)

	if err != nil {
		return "", false
	}

	if time.Now().After(session.ExpiresAt) {
		s.repo.DeleteSession(ctx, token)
		return "", false
	}
	newExpiredAt := time.Now().Add(24 * time.Hour)
	if err :=s.repo.UpdateSessionExpires(ctx, token, newExpiredAt); err != nil {
		log.Printf("Failed to extend session %s: %v", token, err)
	}

	session.ExpiresAt = newExpiredAt
	return session.Username, true
}

func (s *SessionStore) DeleteSession(ctx context.Context, token string) {
	s.repo.DeleteSession(ctx, token)
}

func generateToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (s *SessionStore) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Session-Token")
		ctx := r.Context()

		if token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		username, valid := s.ValidateSession(ctx, token)
		if !valid {
			http.Error(w, "Session expired or invalid", http.StatusUnauthorized)
			return
		}

		ctx = context.WithValue(r.Context(), "username", username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
