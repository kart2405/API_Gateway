package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/kart2405/API_Gateway/internal/config"
	"github.com/kart2405/API_Gateway/internal/middleware/auth"
	"github.com/kart2405/API_Gateway/internal/middleware/ratelimit"
	"github.com/kart2405/API_Gateway/internal/models"
	"github.com/kart2405/API_Gateway/internal/services"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Load route map from YAML config
	config.InitDatabase()
	config.DB.AutoMigrate(&models.User{})
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

		var user models.User
		// Check if user exists
		err := config.DB.Where("username = ?", login.Username).First(&user).Error

		if err != nil {
			// User doesn't exist, create new user
			hashedPassword, hashErr := bcrypt.GenerateFromPassword([]byte(login.Password), bcrypt.DefaultCost)
			if hashErr != nil {
				c.JSON(500, gin.H{"error": "Failed to hash password"})
				return
			}

			user = models.User{
				Username: login.Username,
				Password: string(hashedPassword),
			}

			if createErr := config.DB.Create(&user).Error; createErr != nil {
				c.JSON(500, gin.H{"error": "Failed to create user"})
				return
			}

			log.Printf("New user created: %s with ID: %d", login.Username, user.ID)
		} else {
			// User exists, verify password
			if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login.Password)); err != nil {
				c.JSON(401, gin.H{"error": "Invalid password"})
				return
			}
		}

		// Generate token using user ID
		token, err := auth.GenerateToken(user.ID, user.Username)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(200, gin.H{
			"token":    token,
			"userID":   user.ID,
			"username": user.Username,
			"message":  "Login successful",
		})
	})

	// GET endpoint to view all users
	r.GET("/users", func(c *gin.Context) {
		var users []models.User

		if err := config.DB.Find(&users).Error; err != nil {
			c.JSON(500, gin.H{"error": "Failed to retrieve users"})
			return
		}

		// Don't return password hashes for security
		var userList []gin.H
		for _, user := range users {
			userList = append(userList, gin.H{
				"id":        user.ID,
				"username":  user.Username,
				"createdAt": user.CreatedAt,
			})
		}

		c.JSON(200, gin.H{
			"users": userList,
			"count": len(userList),
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
