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
		roomAddr = "localhost:50052"
	}
	fileAddr := os.Getenv("FILE_SERVICE_ADDR")
	if fileAddr == "" {
		fileAddr = "localhost:50053"
	}

	r := gin.Default()
	r.Use(middleware.CORS())

	// Public (no auth)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "api-gateway"})
	})

	// Public /api/v1 group — no auth middleware (register, login, refresh)
	public := r.Group("/api/v1")

	// Authenticated /api/v1 group
	api := r.Group("/api/v1")
	api.Use(middleware.Auth())

	// Auth routes: public ones on `public`, protected ones on `api`
	authAddr := os.Getenv("AUTH_SERVICE_ADDR")
	if authAddr == "" {
		authAddr = "localhost:50051"
	}
	handlers.RegisterAuthRoutes(public, api, authAddr)

	// Room and File routes (authenticated)
	handlers.RegisterRoomRoutes(api, roomAddr)
	handlers.RegisterFileRoutes(api, fileAddr)

	// Issue routes
	handlers.RegisterIssueRoutes(api)

	// Notification and Activity routes
	notificationAddr := os.Getenv("NOTIFICATION_SERVICE_ADDR")
	if notificationAddr == "" {
		notificationAddr = "localhost:50055"
	}
	activityHandler, err := handlers.NewActivityHandler(notificationAddr)
	if err != nil {
		log.Printf("Failed to connect to notification service for activity: %v", err)
	} else {
		activityHandler.RegisterActivityRoutes(api)
	}

	log.Printf("API Gateway starting on :%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start: %v", err)
	}
}
