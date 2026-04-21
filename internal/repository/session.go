package repository

import (
	"context"
	"tasksite/internal/model"
	"time"
)

type SessionRepository interface {
	CreateSession(ctx context.Context, token, username string, expiresAt time.Time) error
	GetSessionByToken(ctx context.Context, token string) (*model.Session, error)
	UpdateSessionExpires(ctx context.Context, token string, newExpiredAt time.Time) error
	DeleteSession(ctx context.Context, token string) error
}
