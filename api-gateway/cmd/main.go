package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("Starting API Gateway...")

	// Initialize Gin router
	r := gin.Default()

	// Simple health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "API Gateway is running!",
		})
	})

	// TODO: Load config, connect to gRPC clients, add routing
	
	// Start the server on port 8080
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start API Gateway: %v", err)
	}
}
