package domain

import (
	"context"

	"github.com/google/uuid"
)

type AuthUseCase interface {
	Register(ctx context.Context, email, password, fullName string) (*User, error)
	Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error)
	ValidateToken(ctx context.Context, accessToken string) (userID, role string, err error)
	RefreshToken(ctx context.Context, refreshToken string) (newAccess, newRefresh string, err error)
	Logout(ctx context.Context, accessToken string) error
	GetProfile(ctx context.Context, userID uuid.UUID) (*User, error)
}

type UserRepository interface {
	Create(ctx context.Context, user *User) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
}

type SessionRepository interface {
	Create(ctx context.Context, session *Session) (*Session, error)
	GetByRefreshToken(ctx context.Context, token string) (*Session, error)
	GetByAccessToken(ctx context.Context, token string) (*Session, error)
	Revoke(ctx context.Context, sessionID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}
