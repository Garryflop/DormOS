package handlers

import (
	"context"
	"io"
	"net/http"
	"time"

	filev1 "github.com/Garryflop/DormOS-gen-go/file/v1"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type FileHandler struct {
	client filev1.FileServiceClient
}

// RegisterFileRoutes wires all file HTTP endpoints.
func RegisterFileRoutes(r *gin.RouterGroup, fileServiceAddr string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	conn, err := grpc.DialContext(ctx, fileServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		panic("failed to connect to file-service: " + err.Error())
	}
	
	client := filev1.NewFileServiceClient(conn)
	h := &FileHandler{client: client}

	files := r.Group("/files")
	{
		files.POST("/upload", h.UploadFile)    // student+
		files.GET("/:id/url", h.GetFileURL)   // student+
		files.DELETE("/:id", h.DeleteFile)    // admin
	}
}

// POST /api/v1/files/upload
// Accepts multipart/form-data with field "file"
func (h *FileHandler) UploadFile(c *gin.Context) {
	var userIDStr string
	if val, exists := c.Get("user_id"); exists {
		if s, ok := val.(string); ok {
			userIDStr = s
		}
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file field is required"})
		return
	}

	f, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot open file"})
		return
	}
	defer f.Close()

	// 1. Open Client Stream
	stream, err := h.client.UploadFile(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open upload stream: " + err.Error()})
		return
	}

	// 2. Send Metadata first
	err = stream.Send(&filev1.UploadFileRequest{
		Request: &filev1.UploadFileRequest_Metadata{
			Metadata: &filev1.FileMetadata{
				Filename:    fileHeader.Filename,
				ContentType: fileHeader.Header.Get("Content-Type"),
				UploadedBy:  userIDStr,
			},
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send metadata: " + err.Error()})
		return
	}

	// 3. Send file data in chunks
	buffer := make([]byte, 64*1024) // 64KB chunks
	for {
		n, err := f.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file chunk: " + err.Error()})
			return
		}

		err = stream.Send(&filev1.UploadFileRequest{
			Request: &filev1.UploadFileRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send chunk: " + err.Error()})
			return
		}
	}

	// 4. Close stream and receive response
	resp, err := stream.CloseAndRecv()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to finalize upload: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "file uploaded",
		"file_id":      resp.GetFileId(),
		"url":          resp.GetUrl(),
		"filename":     fileHeader.Filename,
		"size":         fileHeader.Size,
		"uploaded_by":  userIDStr,
		"content_type": fileHeader.Header.Get("Content-Type"),
	})
}

// GET /api/v1/files/:id/url
func (h *FileHandler) GetFileURL(c *gin.Context) {
	fileID := c.Param("id")
	
	resp, err := h.client.GetFileURL(c.Request.Context(), &filev1.GetFileURLRequest{
		FileId: fileID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get file URL: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"file_id":      fileID,
		"url":          resp.GetUrl(),
		"filename":     resp.GetFilename(),
		"content_type": resp.GetContentType(),
	})
}

// DELETE /api/v1/files/:id
func (h *FileHandler) DeleteFile(c *gin.Context) {
	fileID := c.Param("id")
	
	resp, err := h.client.DeleteFile(c.Request.Context(), &filev1.DeleteFileRequest{
		FileId: fileID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete file: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "file deleted",
		"file_id": fileID,
		"success": resp.GetSuccess(),
	})
}
