package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Garryflop/DormManage/auth-service/internal/domain"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) domain.UserRepository {
	_, _ = db.Exec(context.Background(), `ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check`)
	_, _ = db.Exec(context.Background(), `ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('student', 'manager', 'admin'))`)
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	query := `
		INSERT INTO users (id, email, password, full_name, phone, role, avatar_url, is_suspended, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, email, full_name, phone, role, avatar_url, is_suspended, created_at, updated_at`

	row := r.db.QueryRow(ctx, query,
		user.ID, user.Email, user.Password, user.FullName, user.Phone,
		string(user.Role), user.AvatarURL, user.IsSuspended, user.CreatedAt, user.UpdatedAt,
	)
	res, err := scanUser(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrEmailTaken
		}
		return nil, err
	}
	return res, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, email, full_name, phone, role, avatar_url, is_suspended, created_at, updated_at
		FROM users WHERE id = $1`
	return scanUser(r.db.QueryRow(ctx, query, id))
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, password, full_name, phone, role, avatar_url, is_suspended, created_at, updated_at
		FROM users WHERE email = $1`
	return scanUserWithPassword(r.db.QueryRow(ctx, query, email))
}

func scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User
	var role string
	err := row.Scan(&u.ID, &u.Email, &u.FullName, &u.Phone, &role, &u.AvatarURL, &u.IsSuspended, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("scan user: %w", err)
	}
	u.Role = domain.Role(role)
	return &u, nil
}

func scanUserWithPassword(row pgx.Row) (*domain.User, error) {
	var u domain.User
	var role string
	err := row.Scan(&u.ID, &u.Email, &u.Password, &u.FullName, &u.Phone, &role, &u.AvatarURL, &u.IsSuspended, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("scan user: %w", err)
	}
	u.Role = domain.Role(role)
	return &u, nil
}
