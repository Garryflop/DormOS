package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

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
	ErrInvalidRole      = errors.New("role must be one of: student, floor_warden, dorm_admin, finance")
	ErrDormIDRequired   = errors.New("dorm_id is required")
	ErrEmailTaken       = errors.New("a user with this email already exists")

	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidPassword = errors.New("invalid password")
	ErrAccountInactive = errors.New("account is inactive")
	ErrInvalidToken    = errors.New("invalid or expired token")
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionRevoked  = errors.New("session has been revoked")
	ErrSessionExpired  = errors.New("session has expired")

	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden: insufficient permissions")
)
