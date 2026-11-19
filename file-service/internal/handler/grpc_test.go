package handler_test

import (
	"context"
	"testing"

	filev1 "github.com/Garryflop/DormOS-gen-go/file/v1"
	"github.com/Garryflop/DormManage/file-service/internal/handler"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ─── GetFileURL Validation Tests ──────────────────────────────────────────────

func TestFileGRPCServer_GetFileURL_EmptyFileId(t *testing.T) {
	srv := handler.NewFileGRPCServer(nil, nil)

	resp, err := srv.GetFileURL(context.Background(), &filev1.GetFileURLRequest{
		FileId: "",
	})

	assert.Nil(t, resp)
	assert.NotNil(t, err)

	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Equal(t, "file_id is required", st.Message())
}

// ─── DeleteFile Validation Tests ──────────────────────────────────────────────

func TestFileGRPCServer_DeleteFile_EmptyFileId(t *testing.T) {
	srv := handler.NewFileGRPCServer(nil, nil)

	resp, err := srv.DeleteFile(context.Background(), &filev1.DeleteFileRequest{
		FileId: "",
	})

	assert.Nil(t, resp)
	assert.NotNil(t, err)

	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Equal(t, "file_id is required", st.Message())
}

// ─── Table-Driven Validation Tests ────────────────────────────────────────────

func TestFileGRPCServer_GetFileURL_Validation_TableDriven(t *testing.T) {
	srv := handler.NewFileGRPCServer(nil, nil)

	tests := []struct {
		name    string
		fileID  string
		wantErr bool
		code    codes.Code
		msg     string
	}{
		{
			name:    "empty file_id returns InvalidArgument",
			fileID:  "",
			wantErr: true,
			code:    codes.InvalidArgument,
			msg:     "file_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := srv.GetFileURL(context.Background(), &filev1.GetFileURLRequest{
				FileId: tt.fileID,
			})

			if tt.wantErr {
				assert.Nil(t, resp)
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.code, st.Code())
				assert.Equal(t, tt.msg, st.Message())
			}
		})
	}
}

func TestFileGRPCServer_DeleteFile_Validation_TableDriven(t *testing.T) {
	srv := handler.NewFileGRPCServer(nil, nil)

	tests := []struct {
		name    string
		fileID  string
		wantErr bool
		code    codes.Code
		msg     string
	}{
		{
			name:    "empty file_id returns InvalidArgument",
			fileID:  "",
			wantErr: true,
			code:    codes.InvalidArgument,
			msg:     "file_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := srv.DeleteFile(context.Background(), &filev1.DeleteFileRequest{
				FileId: tt.fileID,
			})

			if tt.wantErr {
				assert.Nil(t, resp)
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.code, st.Code())
				assert.Equal(t, tt.msg, st.Message())
			}
		})
	}
}
