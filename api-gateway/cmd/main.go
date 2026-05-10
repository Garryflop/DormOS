package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/Garryflop/DormManage/api-gateway/middleware"
)

func main() {
	port := os.Getenv("API_GATEWAY_PORT")
	if port == "" {
		port = "8080"
	}

	r := gin.Default()

	// Global middleware
	r.Use(middleware.CORS())

	// Health check (public)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "api-gateway",
		})
	})

	// Route groups — each member registers their own handlers
	// auth routes registered by D in handlers/auth.go
	// room + file routes registered by N in handlers/room.go + handlers/file.go
	// issue routes registered by E in handlers/issue.go
	// activity routes registered by T in handlers/activity.go

	api := r.Group("/api/v1")
	api.Use(middleware.Auth())
	// TODO: each member registers routes here
	_ = api // suppress unused warning until handlers are added

	log.Printf("API Gateway starting on :%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start: %v", err)
	}
}
