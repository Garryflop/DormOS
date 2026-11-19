package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/Garryflop/DormManage/file-service/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
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

	// Run migration inline (matches 000001_init.up.sql)
	_, err = db.Exec(ctx, `
		CREATE EXTENSION IF NOT EXISTS "pgcrypto";

		CREATE TABLE IF NOT EXISTS files (
			id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			filename     VARCHAR(255) NOT NULL,
			content_type VARCHAR(100) NOT NULL,
			bucket       VARCHAR(100) NOT NULL DEFAULT 'dormos-files',
			object_key   VARCHAR(512) NOT NULL UNIQUE,
			uploaded_by  UUID         NOT NULL,
			created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_files_uploaded_by ON files(uploaded_by);
	`)
	require.NoError(t, err)

	cleanup := func() {
		db.Close()
		pgContainer.Terminate(ctx)
	}

	return db, cleanup
}

// ─── Save Tests ───────────────────────────────────────────────────────────────

func TestFileRepo_Save(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewFileRepository(db)

	record, err := repo.Save(ctx, "photo.jpg", "image/jpeg", "abc123/photo.jpg", "11111111-1111-1111-1111-111111111111")

	assert.NoError(t, err)
	assert.NotEmpty(t, record.ID)
	assert.Equal(t, "photo.jpg", record.Filename)
	assert.Equal(t, "image/jpeg", record.ContentType)
	assert.Equal(t, "abc123/photo.jpg", record.ObjectKey)
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", record.UploadedBy)
	assert.NotZero(t, record.CreatedAt)
}

func TestFileRepo_Save_DuplicateObjectKey(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewFileRepository(db)

	_, err := repo.Save(ctx, "a.txt", "text/plain", "same-key/a.txt", "11111111-1111-1111-1111-111111111111")
	require.NoError(t, err)

	// Second save with same object_key should fail (UNIQUE constraint)
	_, err = repo.Save(ctx, "b.txt", "text/plain", "same-key/a.txt", "22222222-2222-2222-2222-222222222222")
	assert.Error(t, err)
}

// ─── Get Tests ────────────────────────────────────────────────────────────────

func TestFileRepo_Get(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewFileRepository(db)

	saved, err := repo.Save(ctx, "report.pdf", "application/pdf", "uuid1/report.pdf", "11111111-1111-1111-1111-111111111111")
	require.NoError(t, err)

	found, err := repo.Get(ctx, saved.ID)

	assert.NoError(t, err)
	assert.Equal(t, saved.ID, found.ID)
	assert.Equal(t, "report.pdf", found.Filename)
	assert.Equal(t, "application/pdf", found.ContentType)
	assert.Equal(t, "uuid1/report.pdf", found.ObjectKey)
}

func TestFileRepo_Get_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewFileRepository(db)

	_, err := repo.Get(ctx, "00000000-0000-0000-0000-000000000000")
	assert.Error(t, err)
}

// ─── Delete Tests ─────────────────────────────────────────────────────────────

func TestFileRepo_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewFileRepository(db)

	saved, err := repo.Save(ctx, "temp.txt", "text/plain", "uuid2/temp.txt", "11111111-1111-1111-1111-111111111111")
	require.NoError(t, err)

	objectKey, err := repo.Delete(ctx, saved.ID)

	assert.NoError(t, err)
	assert.Equal(t, "uuid2/temp.txt", objectKey)

	// Verify it's actually gone
	_, err = repo.Get(ctx, saved.ID)
	assert.Error(t, err)
}

func TestFileRepo_Delete_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewFileRepository(db)

	_, err := repo.Delete(ctx, "00000000-0000-0000-0000-000000000000")
	assert.Error(t, err)
}

// ─── Multiple Files Test ──────────────────────────────────────────────────────

func TestFileRepo_SaveAndGetMultiple(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewFileRepository(db)

	files := []struct {
		filename    string
		contentType string
		objectKey   string
	}{
		{"doc1.pdf", "application/pdf", "a/doc1.pdf"},
		{"img1.png", "image/png", "b/img1.png"},
		{"data.csv", "text/csv", "c/data.csv"},
	}

	var ids []string
	for _, f := range files {
		rec, err := repo.Save(ctx, f.filename, f.contentType, f.objectKey, "11111111-1111-1111-1111-111111111111")
		require.NoError(t, err)
		ids = append(ids, rec.ID)
	}

	// Verify each file can be retrieved independently
	for i, id := range ids {
		rec, err := repo.Get(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, files[i].filename, rec.Filename)
		assert.Equal(t, files[i].contentType, rec.ContentType)
	}
}
