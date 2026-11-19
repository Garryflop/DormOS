package domain

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusOpen       Status = "open"
	StatusInProgress Status = "in_progress"
	StatusResolved   Status = "resolved"
	StatusClosed     Status = "closed"
)

type Issue struct {
	ID          uuid.UUID  `db:"id"`
	UserID      uuid.UUID  `db:"user_id"`
	RoomNumber  string     `db:"room_number"`
	CategoryID  uuid.UUID  `db:"category_id"`
	Title       string     `db:"title"`
	Description string     `db:"description"`
	Status      Status     `db:"status"`
	WorkerID    *uuid.UUID `db:"worker_id"`
	PhotoURL    *string    `db:"photo_url"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

type Comment struct {
	ID        uuid.UUID `db:"id"`
	IssueID   uuid.UUID `db:"issue_id"`
	UserID    uuid.UUID `db:"user_id"`
	Text      string    `db:"text"`
	CreatedAt time.Time `db:"created_at"`
}

type Worker struct {
	ID        uuid.UUID `db:"id"`
	Name      string    `db:"name"`
	Specialty string    `db:"specialty"`
	Phone     string    `db:"phone"`
	IsActive  bool      `db:"is_active"`
	CreatedAt time.Time `db:"created_at"`
}

type Category struct {
	ID        uuid.UUID `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
}

// Input structs

type CreateIssueInput struct {
	UserID      uuid.UUID
	RoomNumber  string
	CategoryID  uuid.UUID
	Title       string
	Description string
	PhotoURL    *string
}

type UpdateIssueStatusInput struct {
	IssueID  uuid.UUID
	Status   Status
	WorkerID *uuid.UUID
}

type AddCommentInput struct {
	IssueID uuid.UUID
	UserID  uuid.UUID
	Text    string
}
