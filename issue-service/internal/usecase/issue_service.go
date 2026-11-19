package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/Garryflop/DormManage/issue-service/internal/domain"
	issuenats "github.com/Garryflop/DormManage/issue-service/internal/transport/nats"
)

type IssueService struct {
	issues     domain.IssueRepository
	comments   domain.CommentRepository
	workers    domain.WorkerRepository
	categories domain.CategoryRepository
	publisher  *issuenats.Publisher
	redis      *redis.Client
}

func New(
	issues domain.IssueRepository,
	comments domain.CommentRepository,
	workers domain.WorkerRepository,
	categories domain.CategoryRepository,
	publisher *issuenats.Publisher,
	redis *redis.Client,
) *IssueService {
	return &IssueService{
		issues:     issues,
		comments:   comments,
		workers:    workers,
		categories: categories,
		publisher:  publisher,
		redis:      redis,
	}
}

// 1. CreateIssue
func (s *IssueService) CreateIssue(ctx context.Context, in domain.CreateIssueInput) (*domain.Issue, error) {
	issue, err := s.issues.Create(ctx, in)
	if err != nil {
		return nil, err
	}
	if s.publisher != nil {
		s.publisher.PublishIssueCreated(ctx, issuenats.IssueCreatedEvent{
			IssueID:    issue.ID,
			UserID:     issue.UserID,
			RoomNumber: issue.RoomNumber,
			Title:      issue.Title,
			CreatedAt:  issue.CreatedAt,
		})
	}
	return issue, nil
}

// 2. GetIssue
func (s *IssueService) GetIssue(ctx context.Context, id uuid.UUID) (*domain.Issue, error) {
	return s.issues.GetByID(ctx, id)
}

// 3. ListMyIssues
func (s *IssueService) ListMyIssues(ctx context.Context, userID uuid.UUID) ([]domain.Issue, error) {
	return s.issues.ListByUser(ctx, userID)
}

// 4. ListAllIssues
func (s *IssueService) ListAllIssues(ctx context.Context) ([]domain.Issue, error) {
	return s.issues.ListAll(ctx)
}

// 5. UpdateIssueStatus
func (s *IssueService) UpdateIssueStatus(ctx context.Context, in domain.UpdateIssueStatusInput) (*domain.Issue, error) {
	old, err := s.issues.GetByID(ctx, in.IssueID)
	if err != nil {
		return nil, err
	}
	issue, err := s.issues.UpdateStatus(ctx, in)
	if err != nil {
		return nil, err
	}
	if s.publisher != nil {
		s.publisher.PublishIssueStatusChanged(ctx, issuenats.IssueStatusChangedEvent{
			IssueID:   issue.ID,
			UserID:    issue.UserID,
			OldStatus: string(old.Status),
			NewStatus: string(issue.Status),
			ChangedAt: time.Now(),
		})
	}
	return issue, nil
}

// 6. DeleteIssue
func (s *IssueService) DeleteIssue(ctx context.Context, id uuid.UUID) error {
	return s.issues.Delete(ctx, id)
}

// 7. AddComment
func (s *IssueService) AddComment(ctx context.Context, in domain.AddCommentInput) (*domain.Comment, error) {
	return s.comments.Add(ctx, in)
}

// 8. ListComments
func (s *IssueService) ListComments(ctx context.Context, issueID uuid.UUID) ([]domain.Comment, error) {
	return s.comments.ListByIssue(ctx, issueID)
}

// 9. AssignWorker
func (s *IssueService) AssignWorker(ctx context.Context, issueID, workerID uuid.UUID) (*domain.Issue, error) {
	return s.issues.AssignWorker(ctx, issueID, workerID)
}

// 10. ListWorkers
func (s *IssueService) ListWorkers(ctx context.Context) ([]domain.Worker, error) {
	return s.workers.ListActive(ctx)
}

// 11. CreateCategory
func (s *IssueService) CreateCategory(ctx context.Context, name string) (*domain.Category, error) {
	if name == "" {
		return nil, fmt.Errorf("category name is required")
	}
	return s.categories.Create(ctx, name)
}

// 12. ListCategories
func (s *IssueService) ListCategories(ctx context.Context) ([]domain.Category, error) {
	return s.categories.List(ctx)
}
