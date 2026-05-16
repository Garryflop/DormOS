package handlers

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type fileGRPCClient struct {
	conn *grpc.ClientConn
}

func newFileClient(addr string) (*fileGRPCClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}
	return &fileGRPCClient{conn: conn}, nil
}

// RegisterFileRoutes wires all file HTTP endpoints.
// Called from main.go by Nurdaulet.
func RegisterFileRoutes(r *gin.RouterGroup, fileServiceAddr string) {
	client, err := newFileClient(fileServiceAddr)
	if err != nil {
		panic("failed to connect to file-service: " + err.Error())
	}
	h := &FileHandler{client: client}

	files := r.Group("/files")
	{
		files.POST("/upload", h.UploadFile)    // student+
		files.GET("/:id/url", h.GetFileURL)   // student+
		files.DELETE("/:id", h.DeleteFile)    // admin
	}
}

type FileHandler struct {
	client *fileGRPCClient
}

// POST /api/v1/files/upload
// Accepts multipart/form-data with field "file"
func (h *FileHandler) UploadFile(c *gin.Context) {
	userID, _ := c.Get("user_id")

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

	// Read the file bytes
	data, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot read file"})
		return
	}

	// TODO: stream data to file-service via gRPC streaming UploadFile
	_ = data

	c.JSON(http.StatusCreated, gin.H{
		"message":      "file uploaded",
		"filename":     fileHeader.Filename,
		"size":         fileHeader.Size,
		"uploaded_by":  userID,
		"content_type": fileHeader.Header.Get("Content-Type"),
	})
}

// GET /api/v1/files/:id/url
func (h *FileHandler) GetFileURL(c *gin.Context) {
	fileID := c.Param("id")
	// TODO: call h.client.GetFileURL
	c.JSON(http.StatusOK, gin.H{
		"file_id": fileID,
		"url":     "https://minio.example.com/dormos-files/" + fileID,
	})
}

// DELETE /api/v1/files/:id
func (h *FileHandler) DeleteFile(c *gin.Context) {
	fileID := c.Param("id")
	// TODO: call h.client.DeleteFile
	c.JSON(http.StatusOK, gin.H{"message": "file deleted", "file_id": fileID})
}
