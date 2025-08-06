package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kart2405/API_Gateway/internal/config"
	"github.com/kart2405/API_Gateway/internal/middleware/auth"
	"github.com/kart2405/API_Gateway/internal/middleware/cache"
	"github.com/kart2405/API_Gateway/internal/middleware/health"
	"github.com/kart2405/API_Gateway/internal/middleware/logging"
	"github.com/kart2405/API_Gateway/internal/middleware/ratelimit"
	"github.com/kart2405/API_Gateway/internal/middleware/validation"
	"github.com/kart2405/API_Gateway/internal/models"
	"github.com/kart2405/API_Gateway/internal/services"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Load configuration
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	if err := logging.InitLogger(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logging.Sync()

	logger := logging.GetLogger()
	logger.Info("Starting API Gateway")

	// Initialize database
	config.InitDatabase()
	config.DB.AutoMigrate(&models.User{}, &services.RouteConfig{})

	// Initialize Redis
	config.InitRedis()
	defer config.CloseRedis()

	// Initialize route optimizer
	if err := services.InitializeRouteOptimizer(); err != nil {
		logger.Warn("Failed to initialize route optimizer", zap.Error(err))
	} else {
		logger.Info("Route optimizer initialized successfully")
	}

	// Initialize services
	healthService := health.NewHealthService()
	cacheService := cache.NewCacheService()
	validationService := validation.CreateCommonValidationService()

	// Start cache cleanup
	cacheService.StartCleanup(10 * time.Minute)

	// Create Gin router
	r := gin.New()

	// Add middleware
	r.Use(logging.RequestLogger())
	r.Use(logging.LoggingMiddleware())

	// Health check endpoints
	r.Use(health.HealthMiddleware(healthService))

	// Debug endpoint to check config
	r.GET("/debug/config", func(c *gin.Context) {
		appConfig := config.GetConfig()
		c.JSON(200, gin.H{
			"route_map":       config.RouteMap,
			"optimizer_stats": services.GlobalRouteOptimizer.GetRouteStats(),
			"config":          appConfig,
		})
	})

	// Cache statistics endpoint
	r.GET("/debug/cache", func(c *gin.Context) {
		stats := cacheService.GetStats()
		c.JSON(200, stats)
	})

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
	admin.Use(validation.ValidationMiddleware(validationService))
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

	// Protected routes with rate limiting and caching
	protected := r.Group("/")
	protected.Use(auth.JWTAuthMiddleware())
	protected.Use(ratelimit.RateLimitMiddleware())
	protected.Use(cache.CacheMiddleware(cacheService, 5*time.Minute))
	{
		// Reverse Proxy route (uses optimized route lookup)
		protected.Any("/proxy/:service/*proxyPath", services.ReverseProxyHandler)
	}

	// API documentation endpoint
	r.GET("/docs", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"name":        "API Gateway",
			"version":     "1.0.0",
			"description": "High-performance API Gateway with rate limiting, caching, and health monitoring",
			"endpoints": gin.H{
				"health": gin.H{
					"GET /health":      "Complete health status",
					"GET /health/ready": "Readiness check",
					"GET /health/live":  "Liveness check",
				},
				"auth": gin.H{
					"POST /login": "User authentication",
				},
				"admin": gin.H{
					"GET  /admin/routes":                    "List all routes",
					"POST /admin/routes":                    "Create new route",
					"PUT  /admin/routes/:id":                "Update route",
					"DELETE /admin/routes/:id":              "Delete route",
					"GET  /admin/routes/stats":              "Route statistics",
					"GET  /admin/routes/optimizer/stats":    "Optimizer statistics",
					"POST /admin/routes/optimizer/benchmark": "Performance benchmark",
				},
				"proxy": gin.H{
					"ANY /proxy/:service/*proxyPath": "Reverse proxy to backend services",
				},
				"debug": gin.H{
					"GET /debug/config": "Configuration information",
					"GET /debug/cache":  "Cache statistics",
				},
			},
		})
	})

	// Start the gateway server
	port := config.GetConfig().Server.Port
	if port == "" {
		port = "8080"
	}

	logger.Info("Starting server", zap.String("port", port))
	if err := r.Run(":" + port); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
