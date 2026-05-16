package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Garryflop/DormManage/issue-service/internal/domain"
)

type IssueRepo struct {
	db *pgxpool.Pool
}

func NewIssueRepo(db *pgxpool.Pool) *IssueRepo {
	return &IssueRepo{db: db}
}

// 1. CreateIssue
func (r *IssueRepo) Create(ctx context.Context, in domain.CreateIssueInput) (*domain.Issue, error) {
	issue := &domain.Issue{
		ID:          uuid.New(),
		UserID:      in.UserID,
		RoomNumber:  in.RoomNumber,
		CategoryID:  in.CategoryID,
		Title:       in.Title,
		Description: in.Description,
		Status:      domain.StatusOpen,
		PhotoURL:    in.PhotoURL,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	_, err := r.db.Exec(ctx, `
		INSERT INTO issues (id, user_id, room_number, category_id, title, description, status, photo_url, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	`, issue.ID, issue.UserID, issue.RoomNumber, issue.CategoryID,
		issue.Title, issue.Description, issue.Status, issue.PhotoURL,
		issue.CreatedAt, issue.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create issue: %w", err)
	}
	return issue, nil
}

// 2. GetIssue
func (r *IssueRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Issue, error) {
	row := r.db.QueryRow(ctx, `SELECT id, user_id, room_number, category_id, title, description, status, worker_id, photo_url, created_at, updated_at FROM issues WHERE id = $1`, id)
	var i domain.Issue
	if err := row.Scan(&i.ID, &i.UserID, &i.RoomNumber, &i.CategoryID, &i.Title, &i.Description, &i.Status, &i.WorkerID, &i.PhotoURL, &i.CreatedAt, &i.UpdatedAt); err != nil {
		return nil, fmt.Errorf("get issue: %w", err)
	}
	return &i, nil
}

// 3. ListMyIssues
func (r *IssueRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Issue, error) {
	rows, err := r.db.Query(ctx, `SELECT id, user_id, room_number, category_id, title, description, status, worker_id, photo_url, created_at, updated_at FROM issues WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list my issues: %w", err)
	}
	defer rows.Close()
	return scanIssues(rows)
}

// 4. ListAllIssues
func (r *IssueRepo) ListAll(ctx context.Context) ([]domain.Issue, error) {
	rows, err := r.db.Query(ctx, `SELECT id, user_id, room_number, category_id, title, description, status, worker_id, photo_url, created_at, updated_at FROM issues ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list all issues: %w", err)
	}
	defer rows.Close()
	return scanIssues(rows)
}

// 5. UpdateIssueStatus
func (r *IssueRepo) UpdateStatus(ctx context.Context, in domain.UpdateIssueStatusInput) (*domain.Issue, error) {
	_, err := r.db.Exec(ctx, `UPDATE issues SET status=$1, worker_id=$2, updated_at=$3 WHERE id=$4`,
		in.Status, in.WorkerID, time.Now(), in.IssueID)
	if err != nil {
		return nil, fmt.Errorf("update issue status: %w", err)
	}
	return r.GetByID(ctx, in.IssueID)
}

// 6. DeleteIssue
func (r *IssueRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM issues WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete issue: %w", err)
	}
	return nil
}

// 9. AssignWorker
func (r *IssueRepo) AssignWorker(ctx context.Context, issueID, workerID uuid.UUID) (*domain.Issue, error) {
	_, err := r.db.Exec(ctx, `UPDATE issues SET worker_id=$1, status='in_progress', updated_at=$2 WHERE id=$3`,
		workerID, time.Now(), issueID)
	if err != nil {
		return nil, fmt.Errorf("assign worker: %w", err)
	}
	return r.GetByID(ctx, issueID)
}

func scanIssues(rows interface {
	Next() bool
	Scan(...any) error
}) ([]domain.Issue, error) {
	var issues []domain.Issue
	for rows.Next() {
		var i domain.Issue
		if err := rows.Scan(&i.ID, &i.UserID, &i.RoomNumber, &i.CategoryID, &i.Title, &i.Description, &i.Status, &i.WorkerID, &i.PhotoURL, &i.CreatedAt, &i.UpdatedAt); err != nil {
			return nil, err
		}
		issues = append(issues, i)
	}
	return issues, nil
}
