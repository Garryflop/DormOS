package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type FileRecord struct {
	ID          string
	Filename    string
	ContentType string
	ObjectKey   string
	UploadedBy  string
	CreatedAt   int64
}

type FileRepository struct {
	db *pgxpool.Pool
}

func NewFileRepository(db *pgxpool.Pool) *FileRepository {
	return &FileRepository{db: db}
}

func (r *FileRepository) Save(ctx context.Context, filename, contentType, objectKey, uploadedBy string) (*FileRecord, error) {
	rec := &FileRecord{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO files (filename, content_type, object_key, uploaded_by)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, filename, content_type, object_key, uploaded_by, EXTRACT(EPOCH FROM created_at)::bigint`,
		filename, contentType, objectKey, uploadedBy,
	).Scan(&rec.ID, &rec.Filename, &rec.ContentType, &rec.ObjectKey, &rec.UploadedBy, &rec.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("save file record: %w", err)
	}
	return rec, nil
}

func (r *FileRepository) Get(ctx context.Context, fileID string) (*FileRecord, error) {
	rec := &FileRecord{}
	err := r.db.QueryRow(ctx,
		`SELECT id, filename, content_type, object_key, uploaded_by, EXTRACT(EPOCH FROM created_at)::bigint
		 FROM files WHERE id = $1`,
		fileID,
	).Scan(&rec.ID, &rec.Filename, &rec.ContentType, &rec.ObjectKey, &rec.UploadedBy, &rec.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get file record: %w", err)
	}
	return rec, nil
}

func (r *FileRepository) Delete(ctx context.Context, fileID string) (string, error) {
	var objectKey string
	err := r.db.QueryRow(ctx,
		`DELETE FROM files WHERE id = $1 RETURNING object_key`, fileID,
	).Scan(&objectKey)
	if err != nil {
		return "", fmt.Errorf("delete file record: %w", err)
	}
	return objectKey, nil
}
