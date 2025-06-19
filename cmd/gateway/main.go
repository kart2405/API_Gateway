package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/kart2405/API_Gateway/internal/config"
	"github.com/kart2405/API_Gateway/internal/middleware/auth"
	"github.com/kart2405/API_Gateway/internal/middleware/ratelimit"
	"github.com/kart2405/API_Gateway/internal/services"
)

func main() {
	// Load route map from YAML config
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize Redis
	config.InitRedis()

	// Test Redis connection
	_, err := config.RedisClient.Ping(config.Ctx).Result()
	if err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
		log.Println("Rate limiting will not work without Redis")
	} else {
		log.Println("Redis connected successfully")
	}

	// Create Gin router
	r := gin.Default()

	// Public routes
	r.POST("/login", func(c *gin.Context) {
		var login struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&login); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}

		// In a real application, validate credentials against a database
		// For demo purposes, we'll accept any username/password
		userID := auth.GenerateUserID() // Generate unique user ID
		token, err := auth.GenerateToken(userID, login.Username)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(200, gin.H{
			"token":    token,
			"userID":   userID,
			"username": login.Username,
		})
	})

	// Protected routes with rate limiting
	protected := r.Group("/")
	protected.Use(auth.JWTAuthMiddleware())
	protected.Use(ratelimit.RateLimitMiddleware())
	{
		// Reverse Proxy route (uses dynamically loaded config.RouteMap)
		protected.Any("/proxy/:service/*proxyPath", services.ReverseProxyHandler)
	}

	// Start the gateway server on port 8080
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
