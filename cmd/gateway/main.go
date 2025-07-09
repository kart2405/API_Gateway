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
	config.DB.AutoMigrate(&models.User{}, &services.RouteConfig{})
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

	// Initialize route optimizer
	if err := services.InitializeRouteOptimizer(); err != nil {
		log.Printf("Warning: Failed to initialize route optimizer: %v", err)
	} else {
		log.Println("Route optimizer initialized successfully")
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
		// Look up user by username
		if err := config.DB.Where("username = ?", login.Username).First(&user).Error; err != nil {
			c.JSON(401, gin.H{"error": "Invalid username or password"})
			return
		}

		// Compare password hash
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login.Password)); err != nil {
			c.JSON(401, gin.H{"error": "Invalid username or password"})
			return
		}

		// Generate token using real user ID
		token, err := auth.GenerateToken(user.ID, user.Username)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(200, gin.H{
			"token":    token,
			"userID":   user.ID,
			"username": user.Username,
		})
	})

	// Admin routes (protected by JWT)
	admin := r.Group("/admin")
	admin.Use(auth.JWTAuthMiddleware())
	{
		admin.GET("/routes", services.AdminGetRoutes)
		admin.POST("/routes", services.AdminCreateRoute)
		admin.PUT("/routes/:id", services.AdminUpdateRoute)
		admin.DELETE("/routes/:id", services.AdminDeleteRoute)
		admin.GET("/routes/stats", services.AdminGetRouteStats)
		admin.GET("/routes/optimizer/stats", func(c *gin.Context) {
			stats := services.GlobalRouteOptimizer.GetRouteStats()
			c.JSON(200, stats)
		})
		admin.POST("/routes/optimizer/benchmark", func(c *gin.Context) {
			var request struct {
				ServiceNames []string `json:"service_names"`
			}
			if err := c.ShouldBindJSON(&request); err != nil {
				c.JSON(400, gin.H{"error": "Invalid request"})
				return
			}

			benchmark := services.GlobalRouteOptimizer.BenchmarkRouteLookup(request.ServiceNames)
			c.JSON(200, benchmark)
		})
	}

	// Protected routes with rate limiting
	protected := r.Group("/")
	protected.Use(auth.JWTAuthMiddleware())
	protected.Use(ratelimit.RateLimitMiddleware())
	{
		// Reverse Proxy route (uses optimized route lookup)
		protected.Any("/proxy/:service/*proxyPath", services.ReverseProxyHandler)
	}

	// Start the gateway server on port 8080
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
