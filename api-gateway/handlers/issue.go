package handlers

import (
	"context"
	"net/http"
	"time"

	issuev1 "github.com/Garryflop/DormOS-gen-go/issue/v1"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type IssueHandler struct {
	client issuev1.IssueServiceClient
}

func RegisterIssueRoutes(rg *gin.RouterGroup, issueServiceAddr string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, issueServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		panic("failed to connect to issue-service: " + err.Error())
	}

	h := &IssueHandler{client: issuev1.NewIssueServiceClient(conn)}

	issues := rg.Group("/issues")
	{
		// student+
		issues.POST("", h.CreateIssue)
		issues.GET("/my", h.ListMyIssues)
		issues.GET("/:id", h.GetIssue)
		issues.POST("/:id/comments", h.AddComment)
		issues.GET("/:id/comments", h.ListComments)

		// manager+
		issues.GET("", h.ListAllIssues)
		issues.PATCH("/:id/status", h.UpdateIssueStatus)
		issues.PATCH("/:id/assign", h.AssignWorker)
		issues.GET("/workers", h.ListWorkers)

		// admin
		issues.DELETE("/:id", h.DeleteIssue)
		issues.POST("/categories", h.CreateCategory)
	}

	rg.GET("/categories", h.ListCategories)
}

// 1. POST /api/v1/issues
func (h *IssueHandler) CreateIssue(c *gin.Context) {
	var req struct {
		RoomNumber  string   `json:"room_number" binding:"required"`
		CategoryID  string   `json:"category_id" binding:"required"`
		Title       string   `json:"title" binding:"required"`
		Description string   `json:"description" binding:"required"`
		PhotoUrls   []string `json:"photo_urls"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")

	resp, err := h.client.CreateIssue(c.Request.Context(), &issuev1.CreateIssueRequest{
		UserId:      userID.(string),
		RoomNumber:  req.RoomNumber,
		CategoryId:  req.CategoryID,
		Title:       req.Title,
		Description: req.Description,
		PhotoUrls:   req.PhotoUrls,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"issue_id": resp.IssueId})
}

// 2. GET /api/v1/issues/:id
func (h *IssueHandler) GetIssue(c *gin.Context) {
	resp, err := h.client.GetIssue(c.Request.Context(), &issuev1.GetIssueRequest{
		IssueId: c.Param("id"),
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"issue_id":        resp.IssueId,
		"user_id":         resp.UserId,
		"room_number":     resp.RoomNumber,
		"category":        resp.Category,
		"title":           resp.Title,
		"description":     resp.Description,
		"status":          resp.Status,
		"assigned_worker": resp.AssignedWorker,
		"photo_urls":      resp.PhotoUrls,
		"created_at":      resp.CreatedAt,
		"updated_at":      resp.UpdatedAt,
	})
}

// 3. GET /api/v1/issues/my
func (h *IssueHandler) ListMyIssues(c *gin.Context) {
	userID, _ := c.Get("user_id")
	resp, err := h.client.ListMyIssues(c.Request.Context(), &issuev1.ListMyIssuesRequest{
		UserId: userID.(string),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"issues": resp.Issues})
}

// 4. GET /api/v1/issues
func (h *IssueHandler) ListAllIssues(c *gin.Context) {
	resp, err := h.client.ListAllIssues(c.Request.Context(), &issuev1.ListAllIssuesRequest{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"issues": resp.Issues, "total": resp.Total})
}

// 5. PATCH /api/v1/issues/:id/status
func (h *IssueHandler) UpdateIssueStatus(c *gin.Context) {
	var req struct {
		NewStatus string `json:"new_status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.UpdateIssueStatus(c.Request.Context(), &issuev1.UpdateIssueStatusRequest{
		IssueId:   c.Param("id"),
		NewStatus: req.NewStatus,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": resp.Success})
}

// 6. DELETE /api/v1/issues/:id
func (h *IssueHandler) DeleteIssue(c *gin.Context) {
	resp, err := h.client.DeleteIssue(c.Request.Context(), &issuev1.DeleteIssueRequest{
		IssueId: c.Param("id"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": resp.Success})
}

// 7. POST /api/v1/issues/:id/comments
func (h *IssueHandler) AddComment(c *gin.Context) {
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := c.Get("user_id")
	resp, err := h.client.AddComment(c.Request.Context(), &issuev1.AddCommentRequest{
		IssueId: c.Param("id"),
		UserId:  userID.(string),
		Content: req.Content,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"comment_id": resp.CommentId})
}

// 8. GET /api/v1/issues/:id/comments
func (h *IssueHandler) ListComments(c *gin.Context) {
	resp, err := h.client.ListComments(c.Request.Context(), &issuev1.ListCommentsRequest{
		IssueId: c.Param("id"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"comments": resp.Comments})
}

// 9. PATCH /api/v1/issues/:id/assign
func (h *IssueHandler) AssignWorker(c *gin.Context) {
	var req struct {
		WorkerName string `json:"worker_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.AssignWorker(c.Request.Context(), &issuev1.AssignWorkerRequest{
		IssueId:    c.Param("id"),
		WorkerName: req.WorkerName,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": resp.Success})
}

// 10. GET /api/v1/issues/workers
func (h *IssueHandler) ListWorkers(c *gin.Context) {
	resp, err := h.client.ListWorkers(c.Request.Context(), &issuev1.ListWorkersRequest{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"workers": resp.Workers})
}

// 11. POST /api/v1/issues/categories
func (h *IssueHandler) CreateCategory(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.CreateCategory(c.Request.Context(), &issuev1.CreateCategoryRequest{
		Name: req.Name,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"category_id": resp.CategoryId})
}

// 12. GET /api/v1/categories
func (h *IssueHandler) ListCategories(c *gin.Context) {
	resp, err := h.client.ListCategories(c.Request.Context(), &issuev1.ListCategoriesRequest{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"categories": resp.Categories})
}
