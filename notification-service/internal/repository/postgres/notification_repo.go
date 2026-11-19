package postgres

import (
	"context"
	"fmt"

	"github.com/dormos/notification-service/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NotificationRepository handles DB operations for notifications
type NotificationRepository struct {
	db *pgxpool.Pool
}

func NewNotificationRepository(db *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create saves a new notification to the DB
func (r *NotificationRepository) Create(ctx context.Context, n *domain.Notification) (*domain.Notification, error) {
	query := `
		INSERT INTO notifications (user_id, title, message)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`
	row := r.db.QueryRow(ctx, query, n.UserID, n.Title, n.Message)
	if err := row.Scan(&n.ID, &n.CreatedAt); err != nil {
		return nil, fmt.Errorf("NotificationRepository.Create: %w", err)
	}
	return n, nil
}

// ListByUserID returns all notifications for a specific user, newest first
func (r *NotificationRepository) ListByUserID(ctx context.Context, userID string) ([]*domain.Notification, error) {
	query := `
		SELECT id, user_id, title, message, is_read, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("NotificationRepository.ListByUserID: %w", err)
	}
	defer rows.Close()

	var notifs []*domain.Notification
	for rows.Next() {
		n := &domain.Notification{}
		if err := rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Message, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("NotificationRepository.ListByUserID scan: %w", err)
		}
		notifs = append(notifs, n)
	}
	return notifs, nil
}

// MarkAsRead sets is_read = true for a single notification
func (r *NotificationRepository) MarkAsRead(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `UPDATE notifications SET is_read=true WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("NotificationRepository.MarkAsRead: %w", err)
	}
	return nil
}

// MarkAllAsRead sets is_read = true for all notifications of a user
func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, `UPDATE notifications SET is_read=true WHERE user_id=$1 AND is_read=false`, userID)
	if err != nil {
		return fmt.Errorf("NotificationRepository.MarkAllAsRead: %w", err)
	}
	return nil
}

// GetUserEmail retrieves a user's email address by their ID
func (r *NotificationRepository) GetUserEmail(ctx context.Context, userID string) (string, error) {
	var email string
	query := `SELECT email FROM users WHERE id = $1`
	err := r.db.QueryRow(ctx, query, userID).Scan(&email)
	if err != nil {
		return "", err
	}
	return email, nil
}
