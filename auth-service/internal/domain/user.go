package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleStudent Role = "student"
	RoleManager Role = "manager"
	RoleAdmin   Role = "admin"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleStudent, RoleManager, RoleAdmin:
		return true
	}
	return false
}

type User struct {
	ID          uuid.UUID
	Email       string
	Password    string
	FullName    string
	Phone       string
	Role        Role
	AvatarURL   string
	IsSuspended bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Session struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	RefreshToken string
	ExpiresAt    time.Time
	CreatedAt    time.Time
	Revoked      bool
}

func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

func (s *Session) IsValid() bool {
	return !s.Revoked && !s.IsExpired()
}

var (
	ErrEmailRequired    = errors.New("email is required")
	ErrPasswordRequired = errors.New("password is required")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	ErrFullNameRequired = errors.New("full name is required")
	ErrInvalidRole      = errors.New("invalid role")
	ErrEmailTaken       = errors.New("email already exists")
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrAccountSuspended = errors.New("account is suspended")
	ErrInvalidToken     = errors.New("invalid or expired token")
	ErrSessionNotFound  = errors.New("session not found")
	ErrSessionRevoked   = errors.New("session has been revoked")
	ErrSessionExpired   = errors.New("session has expired")
)
