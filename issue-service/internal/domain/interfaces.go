package domain

import (
	"context"

	"github.com/google/uuid"
)

type IssueRepository interface {
	Create(ctx context.Context, in CreateIssueInput) (*Issue, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Issue, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]Issue, error)
	ListAll(ctx context.Context) ([]Issue, error)
	UpdateStatus(ctx context.Context, in UpdateIssueStatusInput) (*Issue, error)
	Delete(ctx context.Context, id uuid.UUID) error
	AssignWorker(ctx context.Context, issueID, workerID uuid.UUID) (*Issue, error)
}

type CommentRepository interface {
	Add(ctx context.Context, in AddCommentInput) (*Comment, error)
	ListByIssue(ctx context.Context, issueID uuid.UUID) ([]Comment, error)
}

type WorkerRepository interface {
	ListActive(ctx context.Context) ([]Worker, error)
}

type CategoryRepository interface {
	Create(ctx context.Context, name string) (*Category, error)
	List(ctx context.Context) ([]Category, error)
}
