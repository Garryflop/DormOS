package main

import (
	"log"
	"os"

	"github.com/Garryflop/DormManage/api-gateway/handlers"
	"github.com/Garryflop/DormManage/api-gateway/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("API_GATEWAY_PORT")
	if port == "" {
		port = "8080"
	}

	roomAddr := os.Getenv("ROOM_SERVICE_ADDR")
	if roomAddr == "" {
		roomAddr = "localhost:50054"
	}
	fileAddr := os.Getenv("FILE_SERVICE_ADDR")
	if fileAddr == "" {
		fileAddr = "localhost:50053"
	}
	issueAddr := os.Getenv("ISSUE_SERVICE_ADDR")
	if issueAddr == "" {
		issueAddr = "localhost:50052"
	}

	r := gin.Default()
	r.Use(middleware.CORS())

	// Public
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "api-gateway"})
	})

	// Authenticated routes
	api := r.Group("/api/v1")
	api.Use(middleware.Auth())

	handlers.RegisterRoomRoutes(api, roomAddr)
	handlers.RegisterFileRoutes(api, fileAddr)
	handlers.RegisterIssueRoutes(api, issueAddr)

	log.Printf("API Gateway starting on :%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start: %v", err)
	}
}
