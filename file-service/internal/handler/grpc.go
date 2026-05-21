package handler

import (
	"bytes"
	"context"
	"io"
	"log"

	filev1 "github.com/Garryflop/DormOS-gen-go/file/v1"
	"github.com/Garryflop/DormManage/file-service/internal/repository"
	"github.com/Garryflop/DormManage/file-service/internal/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FileGRPCServer struct {
	filev1.UnimplementedFileServiceServer
	repo    *repository.FileRepository
	storage *storage.MinioStorage
}

func NewFileGRPCServer(repo *repository.FileRepository, storage *storage.MinioStorage) *FileGRPCServer {
	return &FileGRPCServer{
		repo:    repo,
		storage: storage,
	}
}

// UploadFile implements filev1.FileServiceServer
func (s *FileGRPCServer) UploadFile(stream filev1.FileService_UploadFileServer) error {
	// 1. Read first message which must contain metadata
	firstReq, err := stream.Recv()
	if err != nil {
		if err == io.EOF {
			return status.Error(codes.InvalidArgument, "empty stream")
		}
		return err
	}

	metadata := firstReq.GetMetadata()
	if metadata == nil {
		return status.Error(codes.InvalidArgument, "first message must contain metadata")
	}

	filename := metadata.GetFilename()
	contentType := metadata.GetContentType()
	uploadedBy := metadata.GetUploadedBy()

	if filename == "" {
		return status.Error(codes.InvalidArgument, "filename is required")
	}

	// 2. Read chunks and assemble file
	var buf bytes.Buffer
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("[file-service] Error receiving chunk: %v", err)
			return err
		}
		buf.Write(req.GetChunkData())
	}

	ctx := stream.Context()

	// 3. Upload to MinIO
	objectKey, err := s.storage.Upload(ctx, filename, contentType, &buf, int64(buf.Len()))
	if err != nil {
		log.Printf("[file-service] Error uploading to storage: %v", err)
		return status.Errorf(codes.Internal, "storage upload failed: %v", err)
	}

	// 4. Save record to Database
	record, err := s.repo.Save(ctx, filename, contentType, objectKey, uploadedBy)
	if err != nil {
		// Clean up MinIO object on DB failure
		_ = s.storage.Delete(ctx, objectKey)
		log.Printf("[file-service] Error saving record to DB: %v", err)
		return status.Errorf(codes.Internal, "database save failed: %v", err)
	}

	// 5. Get temporary presigned URL for response
	url, err := s.storage.GetPresignedURL(ctx, objectKey)
	if err != nil {
		// Just log, we can still return fileId
		log.Printf("[file-service] Error generating presigned URL: %v", err)
	}

	return stream.SendAndClose(&filev1.UploadFileResponse{
		FileId: record.ID,
		Url:    url,
	})
}

// GetFileURL implements filev1.FileServiceServer
func (s *FileGRPCServer) GetFileURL(ctx context.Context, req *filev1.GetFileURLRequest) (*filev1.GetFileURLResponse, error) {
	if req.GetFileId() == "" {
		return nil, status.Error(codes.InvalidArgument, "file_id is required")
	}

	record, err := s.repo.Get(ctx, req.GetFileId())
	if err != nil {
		log.Printf("[file-service] File not found: %s", req.GetFileId())
		return nil, status.Errorf(codes.NotFound, "file not found: %v", err)
	}

	url, err := s.storage.GetPresignedURL(ctx, record.ObjectKey)
	if err != nil {
		log.Printf("[file-service] Error generating presigned URL: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to get file url: %v", err)
	}

	return &filev1.GetFileURLResponse{
		Url:         url,
		Filename:    record.Filename,
		ContentType: record.ContentType,
	}, nil
}

// DeleteFile implements filev1.FileServiceServer
func (s *FileGRPCServer) DeleteFile(ctx context.Context, req *filev1.DeleteFileRequest) (*filev1.DeleteFileResponse, error) {
	if req.GetFileId() == "" {
		return nil, status.Error(codes.InvalidArgument, "file_id is required")
	}

	objectKey, err := s.repo.Delete(ctx, req.GetFileId())
	if err != nil {
		log.Printf("[file-service] File not found to delete: %s", req.GetFileId())
		return nil, status.Errorf(codes.NotFound, "file not found: %v", err)
	}

	err = s.storage.Delete(ctx, objectKey)
	if err != nil {
		log.Printf("[file-service] Error deleting from MinIO: %v", err)
		return nil, status.Errorf(codes.Internal, "storage delete failed: %v", err)
	}

	return &filev1.DeleteFileResponse{
		Success: true,
	}, nil
}
