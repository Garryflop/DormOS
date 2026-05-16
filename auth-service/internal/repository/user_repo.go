package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Garryflop/DormManage/auth-service/internal/domain"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	query := `
		INSERT INTO users (
			id, email, password_hash, full_name, role,
			dorm_id, room_number, floor, is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10, $11
		)
		RETURNING id, email, full_name, role, dorm_id, room_number, floor, is_active, created_at, updated_at`

	row := r.db.QueryRow(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.FullName, string(user.Role),
		user.DormID, user.RoomNumber, user.Floor, user.IsActive, user.CreatedAt, user.UpdatedAt,
	)

	var created domain.User
	var role string
	err := row.Scan(
		&created.ID, &created.Email, &created.FullName, &role,
		&created.DormID, &created.RoomNumber, &created.Floor,
		&created.IsActive, &created.CreatedAt, &created.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, domain.ErrEmailTaken
		}
		return nil, fmt.Errorf("create user: %w", err)
	}
	created.Role = domain.Role(role)
	return &created, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, full_name, role,
		       dorm_id, room_number, floor, is_active, created_at, updated_at
		FROM users WHERE id = $1`
	return r.scanUser(r.db.QueryRow(ctx, query, id))
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, full_name, role,
		       dorm_id, room_number, floor, is_active, created_at, updated_at
		FROM users WHERE email = $1`
	return r.scanUser(r.db.QueryRow(ctx, query, email))
}

func (r *UserRepository) scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User
	var role string
	err := row.Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &role,
		&u.DormID, &u.RoomNumber, &u.Floor, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("scan user: %w", err)
	}
	u.Role = domain.Role(role)
	return &u, nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return containsStr(s, "unique constraint") || containsStr(s, "duplicate key")
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
