package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterIssueRoutes(rg *gin.RouterGroup) {
	issues := rg.Group("/issues")
	{
		// student+
		issues.POST("", createIssue)              // 1. CreateIssue
		issues.GET("/:id", getIssue)              // 2. GetIssue
		issues.GET("/my", listMyIssues)           // 3. ListMyIssues
		issues.POST("/:id/comments", addComment)  // 7. AddComment
		issues.GET("/:id/comments", listComments) // 8. ListComments
	}

	// manager+
	mgr := rg.Group("/issues")
	{
		mgr.GET("", listAllIssues)                  // 4. ListAllIssues
		mgr.PATCH("/:id/status", updateIssueStatus) // 5. UpdateIssueStatus
		mgr.PATCH("/:id/assign", assignWorker)      // 9. AssignWorker
		mgr.GET("/workers", listWorkers)            // 10. ListWorkers
	}

	// admin
	admin := rg.Group("/issues")
	{
		admin.DELETE("/:id", deleteIssue)         // 6. DeleteIssue
		admin.POST("/categories", createCategory) // 11. CreateCategory
	}

	// public
	rg.GET("/categories", listCategories) // 12. ListCategories
}

// 1. POST /api/v1/issues
func createIssue(c *gin.Context) {
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
	// TODO: извлечь user_id из JWT (c.Get("user_id"))
	// TODO: вызвать gRPC IssueService.CreateIssue
	c.JSON(http.StatusCreated, gin.H{"message": "issue created", "data": req})
}

// 2. GET /api/v1/issues/:id
func getIssue(c *gin.Context) {
	id := c.Param("id")
	// TODO: gRPC IssueService.GetIssue(id)
	c.JSON(http.StatusOK, gin.H{"issue_id": id})
}

// 3. GET /api/v1/issues/my
func listMyIssues(c *gin.Context) {
	// TODO: извлечь user_id из JWT
	// TODO: gRPC IssueService.ListMyIssues(userID)
	c.JSON(http.StatusOK, gin.H{"issues": []any{}})
}

// 4. GET /api/v1/issues
func listAllIssues(c *gin.Context) {
	// TODO: gRPC IssueService.ListAllIssues()
	c.JSON(http.StatusOK, gin.H{"issues": []any{}})
}

// 5. PATCH /api/v1/issues/:id/status
func updateIssueStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status   string `json:"status" binding:"required"`
		WorkerID string `json:"worker_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// TODO: gRPC IssueService.UpdateIssueStatus
	c.JSON(http.StatusOK, gin.H{"issue_id": id, "status": req.Status})
}

// 6. DELETE /api/v1/issues/:id
func deleteIssue(c *gin.Context) {
	id := c.Param("id")
	// TODO: gRPC IssueService.DeleteIssue
	c.JSON(http.StatusOK, gin.H{"deleted": id})
}

// 7. POST /api/v1/issues/:id/comments
func addComment(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Text string `json:"text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// TODO: gRPC IssueService.AddComment
	c.JSON(http.StatusCreated, gin.H{"issue_id": id, "text": req.Text})
}

// 8. GET /api/v1/issues/:id/comments
func listComments(c *gin.Context) {
	id := c.Param("id")
	// TODO: gRPC IssueService.ListComments
	c.JSON(http.StatusOK, gin.H{"issue_id": id, "comments": []any{}})
}

// 9. PATCH /api/v1/issues/:id/assign
func assignWorker(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		WorkerID string `json:"worker_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// TODO: gRPC IssueService.AssignWorker
	c.JSON(http.StatusOK, gin.H{"issue_id": id, "worker_id": req.WorkerID})
}

// 10. GET /api/v1/issues/workers
func listWorkers(c *gin.Context) {
	// TODO: gRPC IssueService.ListWorkers
	c.JSON(http.StatusOK, gin.H{"workers": []any{}})
}

// 11. POST /api/v1/issues/categories
func createCategory(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// TODO: gRPC IssueService.CreateCategory
	c.JSON(http.StatusCreated, gin.H{"name": req.Name})
}

// 12. GET /api/v1/categories
func listCategories(c *gin.Context) {
	// TODO: gRPC IssueService.ListCategories (с Redis кешем)
	c.JSON(http.StatusOK, gin.H{"categories": []any{}})
}
