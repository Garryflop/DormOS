package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Garryflop/DormManage/issue-service/internal/domain"
	"github.com/Garryflop/DormManage/issue-service/internal/usecase"
)

// ─── Mocks ────────────────────────────────────────────────────────────────────

type mockIssueRepo struct{ mock.Mock }

func (m *mockIssueRepo) Create(ctx context.Context, in domain.CreateIssueInput) (*domain.Issue, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Issue), args.Error(1)
}
func (m *mockIssueRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Issue, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Issue), args.Error(1)
}
func (m *mockIssueRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Issue, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Issue), args.Error(1)
}
func (m *mockIssueRepo) ListAll(ctx context.Context) ([]domain.Issue, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Issue), args.Error(1)
}
func (m *mockIssueRepo) UpdateStatus(ctx context.Context, in domain.UpdateIssueStatusInput) (*domain.Issue, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Issue), args.Error(1)
}
func (m *mockIssueRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockIssueRepo) AssignWorker(ctx context.Context, issueID, workerID uuid.UUID) (*domain.Issue, error) {
	args := m.Called(ctx, issueID, workerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Issue), args.Error(1)
}

type mockCommentRepo struct{ mock.Mock }

func (m *mockCommentRepo) Add(ctx context.Context, in domain.AddCommentInput) (*domain.Comment, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Comment), args.Error(1)
}
func (m *mockCommentRepo) ListByIssue(ctx context.Context, issueID uuid.UUID) ([]domain.Comment, error) {
	args := m.Called(ctx, issueID)
	return args.Get(0).([]domain.Comment), args.Error(1)
}

type mockWorkerRepo struct{ mock.Mock }

func (m *mockWorkerRepo) ListActive(ctx context.Context) ([]domain.Worker, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Worker), args.Error(1)
}

type mockCategoryRepo struct{ mock.Mock }

func (m *mockCategoryRepo) Create(ctx context.Context, name string) (*domain.Category, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Category), args.Error(1)
}
func (m *mockCategoryRepo) List(ctx context.Context) ([]domain.Category, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Category), args.Error(1)
}

// ─── Helper ───────────────────────────────────────────────────────────────────

func newSvc(issues domain.IssueRepository, comments domain.CommentRepository, workers domain.WorkerRepository, categories domain.CategoryRepository) *usecase.IssueService {
	return usecase.New(issues, comments, workers, categories, nil, nil)
}

func fakeIssue() *domain.Issue {
	return &domain.Issue{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		RoomNumber:  "305",
		CategoryID:  uuid.New(),
		Title:       "Broken tap",
		Description: "Water leaking",
		Status:      domain.StatusOpen,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// ─── Tests ────────────────────────────────────────────────────────────────────

func TestCreateIssue_Success(t *testing.T) {
	repo := &mockIssueRepo{}
	expected := fakeIssue()
	input := domain.CreateIssueInput{
		UserID: expected.UserID, RoomNumber: expected.RoomNumber,
		CategoryID: expected.CategoryID, Title: expected.Title, Description: expected.Description,
	}
	repo.On("Create", mock.Anything, input).Return(expected, nil)

	svc := newSvc(repo, &mockCommentRepo{}, &mockWorkerRepo{}, &mockCategoryRepo{})
	result, err := svc.CreateIssue(context.Background(), input)

	assert.NoError(t, err)
	assert.Equal(t, expected.ID, result.ID)
	assert.Equal(t, domain.StatusOpen, result.Status)
	repo.AssertExpectations(t)
}

func TestCreateIssue_RepoError(t *testing.T) {
	repo := &mockIssueRepo{}
	input := domain.CreateIssueInput{Title: "test"}
	repo.On("Create", mock.Anything, input).Return(nil, errors.New("db error"))

	svc := newSvc(repo, &mockCommentRepo{}, &mockWorkerRepo{}, &mockCategoryRepo{})
	result, err := svc.CreateIssue(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetIssue_Success(t *testing.T) {
	repo := &mockIssueRepo{}
	expected := fakeIssue()
	repo.On("GetByID", mock.Anything, expected.ID).Return(expected, nil)

	svc := newSvc(repo, &mockCommentRepo{}, &mockWorkerRepo{}, &mockCategoryRepo{})
	result, err := svc.GetIssue(context.Background(), expected.ID)

	assert.NoError(t, err)
	assert.Equal(t, expected.ID, result.ID)
}

func TestGetIssue_NotFound(t *testing.T) {
	repo := &mockIssueRepo{}
	id := uuid.New()
	repo.On("GetByID", mock.Anything, id).Return(nil, errors.New("not found"))

	svc := newSvc(repo, &mockCommentRepo{}, &mockWorkerRepo{}, &mockCategoryRepo{})
	result, err := svc.GetIssue(context.Background(), id)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestListMyIssues_Success(t *testing.T) {
	repo := &mockIssueRepo{}
	userID := uuid.New()
	issues := []domain.Issue{*fakeIssue(), *fakeIssue()}
	repo.On("ListByUser", mock.Anything, userID).Return(issues, nil)

	svc := newSvc(repo, &mockCommentRepo{}, &mockWorkerRepo{}, &mockCategoryRepo{})
	result, err := svc.ListMyIssues(context.Background(), userID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestListMyIssues_Empty(t *testing.T) {
	repo := &mockIssueRepo{}
	userID := uuid.New()
	repo.On("ListByUser", mock.Anything, userID).Return([]domain.Issue{}, nil)

	svc := newSvc(repo, &mockCommentRepo{}, &mockWorkerRepo{}, &mockCategoryRepo{})
	result, err := svc.ListMyIssues(context.Background(), userID)

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestListAllIssues_Success(t *testing.T) {
	repo := &mockIssueRepo{}
	issues := []domain.Issue{*fakeIssue(), *fakeIssue(), *fakeIssue()}
	repo.On("ListAll", mock.Anything).Return(issues, nil)

	svc := newSvc(repo, &mockCommentRepo{}, &mockWorkerRepo{}, &mockCategoryRepo{})
	result, err := svc.ListAllIssues(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result, 3)
}

func TestUpdateIssueStatus_Success(t *testing.T) {
	repo := &mockIssueRepo{}
	issue := fakeIssue()
	input := domain.UpdateIssueStatusInput{IssueID: issue.ID, Status: domain.StatusInProgress}
	updated := *issue
	updated.Status = domain.StatusInProgress

	repo.On("GetByID", mock.Anything, issue.ID).Return(issue, nil)
	repo.On("UpdateStatus", mock.Anything, input).Return(&updated, nil)

	svc := newSvc(repo, &mockCommentRepo{}, &mockWorkerRepo{}, &mockCategoryRepo{})
	result, err := svc.UpdateIssueStatus(context.Background(), input)

	assert.NoError(t, err)
	assert.Equal(t, domain.StatusInProgress, result.Status)
}

func TestUpdateIssueStatus_NotFound(t *testing.T) {
	repo := &mockIssueRepo{}
	id := uuid.New()
	input := domain.UpdateIssueStatusInput{IssueID: id, Status: domain.StatusResolved}
	repo.On("GetByID", mock.Anything, id).Return(nil, errors.New("not found"))

	svc := newSvc(repo, &mockCommentRepo{}, &mockWorkerRepo{}, &mockCategoryRepo{})
	result, err := svc.UpdateIssueStatus(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestDeleteIssue_Success(t *testing.T) {
	repo := &mockIssueRepo{}
	id := uuid.New()
	repo.On("Delete", mock.Anything, id).Return(nil)

	svc := newSvc(repo, &mockCommentRepo{}, &mockWorkerRepo{}, &mockCategoryRepo{})
	err := svc.DeleteIssue(context.Background(), id)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDeleteIssue_Error(t *testing.T) {
	repo := &mockIssueRepo{}
	id := uuid.New()
	repo.On("Delete", mock.Anything, id).Return(errors.New("db error"))

	svc := newSvc(repo, &mockCommentRepo{}, &mockWorkerRepo{}, &mockCategoryRepo{})
	err := svc.DeleteIssue(context.Background(), id)

	assert.Error(t, err)
}

func TestAddComment_Success(t *testing.T) {
	repo := &mockCommentRepo{}
	input := domain.AddCommentInput{IssueID: uuid.New(), UserID: uuid.New(), Text: "Fixed tomorrow"}
	expected := &domain.Comment{ID: uuid.New(), IssueID: input.IssueID, UserID: input.UserID, Text: input.Text, CreatedAt: time.Now()}
	repo.On("Add", mock.Anything, input).Return(expected, nil)

	svc := newSvc(&mockIssueRepo{}, repo, &mockWorkerRepo{}, &mockCategoryRepo{})
	result, err := svc.AddComment(context.Background(), input)

	assert.NoError(t, err)
	assert.Equal(t, "Fixed tomorrow", result.Text)
}

func TestListComments_Success(t *testing.T) {
	repo := &mockCommentRepo{}
	issueID := uuid.New()
	comments := []domain.Comment{
		{ID: uuid.New(), IssueID: issueID, Text: "comment 1"},
		{ID: uuid.New(), IssueID: issueID, Text: "comment 2"},
	}
	repo.On("ListByIssue", mock.Anything, issueID).Return(comments, nil)

	svc := newSvc(&mockIssueRepo{}, repo, &mockWorkerRepo{}, &mockCategoryRepo{})
	result, err := svc.ListComments(context.Background(), issueID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestAssignWorker_Success(t *testing.T) {
	repo := &mockIssueRepo{}
	issue := fakeIssue()
	workerID := uuid.New()
	assigned := *issue
	assigned.WorkerID = &workerID

	repo.On("AssignWorker", mock.Anything, issue.ID, workerID).Return(&assigned, nil)

	svc := newSvc(repo, &mockCommentRepo{}, &mockWorkerRepo{}, &mockCategoryRepo{})
	result, err := svc.AssignWorker(context.Background(), issue.ID, workerID)

	assert.NoError(t, err)
	assert.Equal(t, &workerID, result.WorkerID)
}

func TestListWorkers_Success(t *testing.T) {
	repo := &mockWorkerRepo{}
	workers := []domain.Worker{
		{ID: uuid.New(), Name: "Ivan", Specialty: "Plumbing"},
		{ID: uuid.New(), Name: "Petr", Specialty: "Electrical"},
	}
	repo.On("ListActive", mock.Anything).Return(workers, nil)

	svc := newSvc(&mockIssueRepo{}, &mockCommentRepo{}, repo, &mockCategoryRepo{})
	result, err := svc.ListWorkers(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Ivan", result[0].Name)
}

func TestCreateCategory_Success(t *testing.T) {
	repo := &mockCategoryRepo{}
	expected := &domain.Category{ID: uuid.New(), Name: "Plumbing"}
	repo.On("Create", mock.Anything, "Plumbing").Return(expected, nil)

	svc := newSvc(&mockIssueRepo{}, &mockCommentRepo{}, &mockWorkerRepo{}, repo)
	result, err := svc.CreateCategory(context.Background(), "Plumbing")

	assert.NoError(t, err)
	assert.Equal(t, "Plumbing", result.Name)
}

func TestCreateCategory_EmptyName(t *testing.T) {
	svc := newSvc(&mockIssueRepo{}, &mockCommentRepo{}, &mockWorkerRepo{}, &mockCategoryRepo{})
	result, err := svc.CreateCategory(context.Background(), "")

	assert.Error(t, err)
	assert.EqualError(t, err, "category name is required")
	assert.Nil(t, result)
}

func TestListCategories_Success(t *testing.T) {
	repo := &mockCategoryRepo{}
	cats := []domain.Category{
		{ID: uuid.New(), Name: "Plumbing"},
		{ID: uuid.New(), Name: "Electrical"},
		{ID: uuid.New(), Name: "Furniture"},
	}
	repo.On("List", mock.Anything).Return(cats, nil)

	svc := newSvc(&mockIssueRepo{}, &mockCommentRepo{}, &mockWorkerRepo{}, repo)
	result, err := svc.ListCategories(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result, 3)
}
