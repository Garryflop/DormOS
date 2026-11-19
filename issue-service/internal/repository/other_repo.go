package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Garryflop/DormManage/issue-service/internal/domain"
)

type CommentRepo struct {
	db *pgxpool.Pool
}

func NewCommentRepo(db *pgxpool.Pool) *CommentRepo {
	return &CommentRepo{db: db}
}

// 7. AddComment
func (r *CommentRepo) Add(ctx context.Context, in domain.AddCommentInput) (*domain.Comment, error) {
	c := &domain.Comment{
		ID:        uuid.New(),
		IssueID:   in.IssueID,
		UserID:    in.UserID,
		Text:      in.Text,
		CreatedAt: time.Now(),
	}
	_, err := r.db.Exec(ctx,
		`INSERT INTO issue_comments (id, issue_id, user_id, text, created_at) VALUES ($1,$2,$3,$4,$5)`,
		c.ID, c.IssueID, c.UserID, c.Text, c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("add comment: %w", err)
	}
	return c, nil
}

// 8. ListComments
func (r *CommentRepo) ListByIssue(ctx context.Context, issueID uuid.UUID) ([]domain.Comment, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, issue_id, user_id, text, created_at FROM issue_comments WHERE issue_id=$1 ORDER BY created_at ASC`,
		issueID)
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}
	defer rows.Close()

	var comments []domain.Comment
	for rows.Next() {
		var c domain.Comment
		if err := rows.Scan(&c.ID, &c.IssueID, &c.UserID, &c.Text, &c.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}

// --- Workers ---

type WorkerRepo struct {
	db *pgxpool.Pool
}

func NewWorkerRepo(db *pgxpool.Pool) *WorkerRepo {
	return &WorkerRepo{db: db}
}

// 10. ListWorkers
func (r *WorkerRepo) ListActive(ctx context.Context) ([]domain.Worker, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, specialty, phone, is_active, created_at FROM workers WHERE is_active=true ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list workers: %w", err)
	}
	defer rows.Close()

	var workers []domain.Worker
	for rows.Next() {
		var w domain.Worker
		if err := rows.Scan(&w.ID, &w.Name, &w.Specialty, &w.Phone, &w.IsActive, &w.CreatedAt); err != nil {
			return nil, err
		}
		workers = append(workers, w)
	}
	return workers, nil
}

// --- Categories ---

type CategoryRepo struct {
	db *pgxpool.Pool
}

func NewCategoryRepo(db *pgxpool.Pool) *CategoryRepo {
	return &CategoryRepo{db: db}
}

// 11. CreateCategory
func (r *CategoryRepo) Create(ctx context.Context, name string) (*domain.Category, error) {
	cat := &domain.Category{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: time.Now(),
	}
	_, err := r.db.Exec(ctx,
		`INSERT INTO issue_categories (id, name, created_at) VALUES ($1,$2,$3)`,
		cat.ID, cat.Name, cat.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create category: %w", err)
	}
	return cat, nil
}

// 12. ListCategories
func (r *CategoryRepo) List(ctx context.Context) ([]domain.Category, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, created_at FROM issue_categories ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()

	var cats []domain.Category
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.CreatedAt); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, nil
}
