package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioStorage struct {
	client *minio.Client
	bucket string
}

func NewMinioStorage(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*MinioStorage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio init: %w", err)
	}

	// Ensure bucket exists
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("minio bucket check: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("minio make bucket: %w", err)
		}
	}

	return &MinioStorage{client: client, bucket: bucket}, nil
}

// Upload stores the file and returns the object key
func (s *MinioStorage) Upload(ctx context.Context, filename, contentType string, reader io.Reader, size int64) (string, error) {
	objectKey := fmt.Sprintf("%s/%s", uuid.NewString(), filename)
	_, err := s.client.PutObject(ctx, s.bucket, objectKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("minio upload: %w", err)
	}
	return objectKey, nil
}

// GetPresignedURL returns a temporary URL valid for 1 hour
func (s *MinioStorage) GetPresignedURL(ctx context.Context, objectKey string) (string, error) {
	url, err := s.client.PresignedGetObject(ctx, s.bucket, objectKey, time.Hour, nil)
	if err != nil {
		return "", fmt.Errorf("minio presign: %w", err)
	}
	return url.String(), nil
}

// Delete removes a file from MinIO
func (s *MinioStorage) Delete(ctx context.Context, objectKey string) error {
	return s.client.RemoveObject(ctx, s.bucket, objectKey, minio.RemoveObjectOptions{})
}
