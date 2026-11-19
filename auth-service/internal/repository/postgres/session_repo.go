package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Garryflop/DormManage/auth-service/internal/domain"
)

type SessionRepository struct {
	db *pgxpool.Pool
}

func NewSessionRepository(db *pgxpool.Pool) domain.SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, session *domain.Session) (*domain.Session, error) {
	query := `
		INSERT INTO sessions (id, user_id, refresh_token, expires_at, created_at, revoked)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, refresh_token, expires_at, created_at, revoked`
	row := r.db.QueryRow(ctx, query,
		session.ID, session.UserID, session.RefreshToken,
		session.ExpiresAt, session.CreatedAt, session.Revoked,
	)
	return scanSession(row)
}

func (r *SessionRepository) GetByRefreshToken(ctx context.Context, token string) (*domain.Session, error) {
	query := `SELECT id, user_id, refresh_token, expires_at, created_at, revoked FROM sessions WHERE refresh_token = $1`
	return scanSession(r.db.QueryRow(ctx, query, token))
}

func (r *SessionRepository) GetByAccessToken(ctx context.Context, token string) (*domain.Session, error) {
	query := `SELECT id, user_id, refresh_token, expires_at, created_at, revoked FROM sessions WHERE access_token = $1`
	return scanSession(r.db.QueryRow(ctx, query, token))
}

func (r *SessionRepository) Revoke(ctx context.Context, sessionID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE sessions SET revoked = true WHERE id = $1`, sessionID)
	return err
}

func (r *SessionRepository) DeleteExpired(ctx context.Context) error {
	_, err := r.db.Exec(ctx, `DELETE FROM sessions WHERE expires_at < NOW()`)
	return err
}

func scanSession(row pgx.Row) (*domain.Session, error) {
	var s domain.Session
	err := row.Scan(&s.ID, &s.UserID, &s.RefreshToken, &s.ExpiresAt, &s.CreatedAt, &s.Revoked)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSessionNotFound
		}
		return nil, fmt.Errorf("scan session: %w", err)
	}
	return &s, nil
}
