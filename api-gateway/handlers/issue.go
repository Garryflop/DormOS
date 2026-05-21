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

type frontendIssue struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	RoomNumber  string `json:"room_number"`
	CategoryID  string `json:"category_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	WorkerID    string `json:"worker_id,omitempty"`
	PhotoURL    string `json:"photo_url,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type frontendCategory struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type frontendWorker struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Specialty string `json:"specialty"`
}

func mapIssueToFrontend(i *issuev1.GetIssueResponse) frontendIssue {
	var photoURL string
	if len(i.PhotoUrls) > 0 {
		photoURL = i.PhotoUrls[0]
	}
	return frontendIssue{
		ID:          i.IssueId,
		UserID:      i.UserId,
		RoomNumber:  i.RoomNumber,
		CategoryID:  i.Category, // Category contains the Category UUID in gRPC
		Title:       i.Title,
		Description: i.Description,
		Status:      i.Status,
		WorkerID:    i.AssignedWorker,
		PhotoURL:    photoURL,
		CreatedAt:   time.Unix(i.CreatedAt, 0).Format(time.RFC3339),
		UpdatedAt:   time.Unix(i.UpdatedAt, 0).Format(time.RFC3339),
	}
}

// RegisterIssueRoutes wires all issue HTTP endpoints to the real gRPC IssueService.
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
	}

	// manager+
	mgr := rg.Group("/issues")
	{
		mgr.GET("", h.ListAllIssues)
		mgr.PATCH("/:id/status", h.UpdateIssueStatus)
		mgr.PATCH("/:id/assign", h.AssignWorker)
		mgr.GET("/workers", h.ListWorkers)
	}

	// admin
	admin := rg.Group("/issues")
	{
		admin.DELETE("/:id", h.DeleteIssue)
		admin.POST("/categories", h.CreateCategory)
	}

	// public
	rg.GET("/categories", h.ListCategories)
}

// POST /api/v1/issues
func (h *IssueHandler) CreateIssue(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var req struct {
		RoomNumber  string `json:"room_number" binding:"required"`
		CategoryID  string `json:"category_id" binding:"required"`
		Title       string `json:"title" binding:"required"`
		Description string `json:"description" binding:"required"`
		PhotoURL    string `json:"photo_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.CreateIssue(c.Request.Context(), &issuev1.CreateIssueRequest{
		UserId:     userID.(string),
		RoomNumber: req.RoomNumber,
		CategoryId: req.CategoryID,
		Title:      req.Title,
		Description: req.Description,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"issue_id": resp.IssueId})
}

// GET /api/v1/issues/:id
func (h *IssueHandler) GetIssue(c *gin.Context) {
	resp, err := h.client.GetIssue(c.Request.Context(), &issuev1.GetIssueRequest{
		IssueId: c.Param("id"),
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, mapIssueToFrontend(resp))
}

// GET /api/v1/issues/my
func (h *IssueHandler) ListMyIssues(c *gin.Context) {
	userID, _ := c.Get("user_id")
	resp, err := h.client.ListMyIssues(c.Request.Context(), &issuev1.ListMyIssuesRequest{
		UserId: userID.(string),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	issues := make([]frontendIssue, len(resp.Issues))
	for i, item := range resp.Issues {
		issues[i] = mapIssueToFrontend(item)
	}
	c.JSON(http.StatusOK, gin.H{"issues": issues})
}

// GET /api/v1/issues
func (h *IssueHandler) ListAllIssues(c *gin.Context) {
	resp, err := h.client.ListAllIssues(c.Request.Context(), &issuev1.ListAllIssuesRequest{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	issues := make([]frontendIssue, len(resp.Issues))
	for i, item := range resp.Issues {
		issues[i] = mapIssueToFrontend(item)
	}
	c.JSON(http.StatusOK, gin.H{"issues": issues})
}

// PATCH /api/v1/issues/:id/status
func (h *IssueHandler) UpdateIssueStatus(c *gin.Context) {
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.UpdateIssueStatus(c.Request.Context(), &issuev1.UpdateIssueStatusRequest{
		IssueId:   c.Param("id"),
		NewStatus: req.Status,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": resp.Success})
}

// DELETE /api/v1/issues/:id
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

// POST /api/v1/issues/:id/comments
func (h *IssueHandler) AddComment(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var req struct {
		Text string `json:"text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.AddComment(c.Request.Context(), &issuev1.AddCommentRequest{
		IssueId: c.Param("id"),
		UserId:  userID.(string),
		Content: req.Text,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"comment_id": resp.CommentId})
}

// GET /api/v1/issues/:id/comments
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

// PATCH /api/v1/issues/:id/assign
func (h *IssueHandler) AssignWorker(c *gin.Context) {
	var req struct {
		WorkerID string `json:"worker_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.AssignWorker(c.Request.Context(), &issuev1.AssignWorkerRequest{
		IssueId:    c.Param("id"),
		WorkerName: req.WorkerID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": resp.Success})
}

// GET /api/v1/issues/workers
func (h *IssueHandler) ListWorkers(c *gin.Context) {
	resp, err := h.client.ListWorkers(c.Request.Context(), &issuev1.ListWorkersRequest{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	workers := make([]frontendWorker, len(resp.Workers))
	for i, w := range resp.Workers {
		workers[i] = frontendWorker{
			ID:        w.WorkerId,
			Name:      w.Name,
			Specialty: w.Specialization,
		}
	}
	c.JSON(http.StatusOK, gin.H{"workers": workers})
}

// POST /api/v1/issues/categories
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

// GET /api/v1/categories
func (h *IssueHandler) ListCategories(c *gin.Context) {
	resp, err := h.client.ListCategories(c.Request.Context(), &issuev1.ListCategoriesRequest{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	categories := make([]frontendCategory, len(resp.Categories))
	for i, cat := range resp.Categories {
		categories[i] = frontendCategory{
			ID:   cat.CategoryId,
			Name: cat.Name,
		}
	}
	c.JSON(http.StatusOK, gin.H{"categories": categories})
}
