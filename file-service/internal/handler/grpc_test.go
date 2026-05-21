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

func TestFileGRPCServer_GetFileURL_ValidationError(t *testing.T) {
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

func TestFileGRPCServer_DeleteFile_ValidationError(t *testing.T) {
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
