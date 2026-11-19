package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			// Allow any localhost origin for development (e.g., http://localhost:3000, http://localhost:3001, http://localhost:5173)
			return origin == "" || 
				(len(origin) >= 16 && origin[:16] == "http://localhost") || 
				(len(origin) >= 17 && origin[:17] == "https://localhost")
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
