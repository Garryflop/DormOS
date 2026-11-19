package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/Garryflop/DormManage/issue-service/internal/domain"
	"github.com/Garryflop/DormManage/issue-service/internal/repository"
)

func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	ctx := context.Background()

	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)

	// Run migrations
	_, err = db.Exec(ctx, `
		CREATE EXTENSION IF NOT EXISTS "pgcrypto";

		CREATE TABLE IF NOT EXISTS issue_categories (
			id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
			name       VARCHAR(100) NOT NULL UNIQUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS workers (
			id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
			name       VARCHAR(255) NOT NULL,
			specialty  VARCHAR(100) NOT NULL,
			phone      VARCHAR(30),
			is_active  BOOLEAN     NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS issues (
			id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id     UUID         NOT NULL,
			room_number VARCHAR(20)  NOT NULL,
			category_id UUID         NOT NULL REFERENCES issue_categories(id),
			title       VARCHAR(255) NOT NULL,
			description TEXT         NOT NULL,
			status      VARCHAR(50)  NOT NULL DEFAULT 'open',
			worker_id   UUID         NULL REFERENCES workers(id),
			photo_url   VARCHAR(1024) NULL,
			created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
			updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS issue_comments (
			id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
			issue_id   UUID        NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
			user_id    UUID        NOT NULL,
			text       TEXT        NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	require.NoError(t, err)

	cleanup := func() {
		db.Close()
		pgContainer.Terminate(ctx)
	}

	return db, cleanup
}

func TestIssueRepo_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	catRepo := repository.NewCategoryRepo(db)
	issueRepo := repository.NewIssueRepo(db)

	// Create category first
	cat, err := catRepo.Create(ctx, "Plumbing")
	require.NoError(t, err)

	input := domain.CreateIssueInput{
		UserID:      uuid.New(),
		RoomNumber:  "305",
		CategoryID:  cat.ID,
		Title:       "Broken tap",
		Description: "Water leaking",
	}

	issue, err := issueRepo.Create(ctx, input)

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, issue.ID)
	assert.Equal(t, "Broken tap", issue.Title)
	assert.Equal(t, domain.StatusOpen, issue.Status)
	assert.Equal(t, "305", issue.RoomNumber)
}

func TestIssueRepo_GetByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	catRepo := repository.NewCategoryRepo(db)
	issueRepo := repository.NewIssueRepo(db)

	cat, _ := catRepo.Create(ctx, "Electrical")
	input := domain.CreateIssueInput{
		UserID: uuid.New(), RoomNumber: "101",
		CategoryID: cat.ID, Title: "No power", Description: "Power outage",
	}
	created, err := issueRepo.Create(ctx, input)
	require.NoError(t, err)

	found, err := issueRepo.GetByID(ctx, created.ID)

	assert.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "No power", found.Title)
}

func TestIssueRepo_GetByID_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	issueRepo := repository.NewIssueRepo(db)

	_, err := issueRepo.GetByID(ctx, uuid.New())
	assert.Error(t, err)
}

func TestIssueRepo_ListByUser(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	catRepo := repository.NewCategoryRepo(db)
	issueRepo := repository.NewIssueRepo(db)

	cat, _ := catRepo.Create(ctx, "Furniture")
	userID := uuid.New()

	// Create 3 issues for same user
	for i := 0; i < 3; i++ {
		_, err := issueRepo.Create(ctx, domain.CreateIssueInput{
			UserID: userID, RoomNumber: "202",
			CategoryID: cat.ID, Title: "Issue", Description: "desc",
		})
		require.NoError(t, err)
	}

	// Create 1 issue for different user
	_, err := issueRepo.Create(ctx, domain.CreateIssueInput{
		UserID: uuid.New(), RoomNumber: "203",
		CategoryID: cat.ID, Title: "Other", Description: "desc",
	})
	require.NoError(t, err)

	issues, err := issueRepo.ListByUser(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, issues, 3)
}

func TestIssueRepo_UpdateStatus(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	catRepo := repository.NewCategoryRepo(db)
	issueRepo := repository.NewIssueRepo(db)

	cat, _ := catRepo.Create(ctx, "Heating")
	issue, err := issueRepo.Create(ctx, domain.CreateIssueInput{
		UserID: uuid.New(), RoomNumber: "301",
		CategoryID: cat.ID, Title: "Cold room", Description: "No heating",
	})
	require.NoError(t, err)
	assert.Equal(t, domain.StatusOpen, issue.Status)

	updated, err := issueRepo.UpdateStatus(ctx, domain.UpdateIssueStatusInput{
		IssueID: issue.ID,
		Status:  domain.StatusInProgress,
	})

	assert.NoError(t, err)
	assert.Equal(t, domain.StatusInProgress, updated.Status)
}

func TestIssueRepo_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	catRepo := repository.NewCategoryRepo(db)
	issueRepo := repository.NewIssueRepo(db)

	cat, _ := catRepo.Create(ctx, "Internet")
	issue, err := issueRepo.Create(ctx, domain.CreateIssueInput{
		UserID: uuid.New(), RoomNumber: "401",
		CategoryID: cat.ID, Title: "No wifi", Description: "Internet down",
	})
	require.NoError(t, err)

	err = issueRepo.Delete(ctx, issue.ID)
	assert.NoError(t, err)

	_, err = issueRepo.GetByID(ctx, issue.ID)
	assert.Error(t, err)
}

func TestCommentRepo_AddAndList(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	catRepo := repository.NewCategoryRepo(db)
	issueRepo := repository.NewIssueRepo(db)
	commentRepo := repository.NewCommentRepo(db)

	cat, _ := catRepo.Create(ctx, "Security")
	issue, err := issueRepo.Create(ctx, domain.CreateIssueInput{
		UserID: uuid.New(), RoomNumber: "501",
		CategoryID: cat.ID, Title: "Broken lock", Description: "Door lock broken",
	})
	require.NoError(t, err)

	userID := uuid.New()
	comment, err := commentRepo.Add(ctx, domain.AddCommentInput{
		IssueID: issue.ID, UserID: userID, Text: "Will fix tomorrow",
	})

	assert.NoError(t, err)
	assert.Equal(t, "Will fix tomorrow", comment.Text)

	comments, err := commentRepo.ListByIssue(ctx, issue.ID)
	assert.NoError(t, err)
	assert.Len(t, comments, 1)
	assert.Equal(t, "Will fix tomorrow", comments[0].Text)
}

func TestCategoryRepo_CreateAndList(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	catRepo := repository.NewCategoryRepo(db)

	names := []string{"Plumbing", "Electrical", "Furniture"}
	for _, name := range names {
		_, err := catRepo.Create(ctx, name)
		require.NoError(t, err)
	}

	cats, err := catRepo.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, cats, 3)
}
