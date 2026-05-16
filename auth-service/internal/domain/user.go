package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleStudent     Role = "student"
	RoleFloorWarden Role = "floor_warden"
	RoleDormAdmin   Role = "dorm_admin"
	RoleFinance     Role = "finance"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleStudent, RoleFloorWarden, RoleDormAdmin, RoleFinance:
		return true
	}
	return false
}

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	FullName     string
	Role         Role
	DormID       uuid.UUID
	RoomNumber   string
	Floor        int
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type RegisterInput struct {
	Email      string
	Password   string
	FullName   string
	Role       Role
	DormID     uuid.UUID
	RoomNumber string
	Floor      int
}

func (r *RegisterInput) Validate() error {
	if r.Email == "" {
		return ErrEmailRequired
	}
	if r.Password == "" {
		return ErrPasswordRequired
	}
	if len(r.Password) < 8 {
		return ErrPasswordTooShort
	}
	if r.FullName == "" {
		return ErrFullNameRequired
	}
	if !r.Role.IsValid() {
		return ErrInvalidRole
	}
	if r.DormID == uuid.Nil {
		return ErrDormIDRequired
	}
	return nil
}
