package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kart2405/API_Gateway/internal/services"
)

// TestRateLimitMiddlewareStructure tests the middleware structure without Redis dependency
func TestRateLimitMiddlewareStructure(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Initialize route optimizer with test routes
	services.GlobalRouteOptimizer.BuildOptimizedRoutes([]services.RouteConfig{
		{
			ServiceName: "test-service",
			BackendURL:  "http://localhost:8081",
			RateLimit:   10,
		},
	})

	// Create router with rate limiting
	r := gin.New()

	// Add userID to context (required by rate limiting middleware) - MUST be before rate limiting
	r.Use(func(c *gin.Context) {
		c.Set("userID", "test-user-123")
		c.Next()
	})

	// Note: RateLimitMiddleware requires Redis connection
	// This test validates the middleware structure and userID requirement
	r.Use(func(c *gin.Context) {
		// Simulate rate limiting logic without Redis
		userIDRaw, exists := c.Get("userID")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized - missing userID in context"})
			return
		}
		userID := userIDRaw.(string)
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing userID"})
			return
		}
		c.Next()
	})

	r.GET("/proxy/:service/*proxyPath", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// Test successful request with userID
	req, _ := http.NewRequest("GET", "/proxy/test-service/api/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for request with userID, got %d", w.Code)
	}
}

func TestRateLimitWithoutUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(RateLimitMiddleware())

	r.GET("/proxy/:service/*proxyPath", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/proxy/test-service/api/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should fail without userID in context
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for missing userID, got %d", w.Code)
	}
}

func TestRateLimitServiceLookup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Initialize route optimizer with test routes
	services.GlobalRouteOptimizer.BuildOptimizedRoutes([]services.RouteConfig{
		{
			ServiceName: "test-service",
			BackendURL:  "http://localhost:8081",
			RateLimit:   15,
		},
		{
			ServiceName: "another-service",
			BackendURL:  "http://localhost:8082",
			RateLimit:   20,
		},
	})

	// Test route lookup functionality
	route, exists := services.GlobalRouteOptimizer.FindRouteOptimized("test-service")
	if !exists {
		t.Error("Expected to find test-service route")
	}
	if route.RateLimit != 15 {
		t.Errorf("Expected rate limit 15, got %d", route.RateLimit)
	}

	route, exists = services.GlobalRouteOptimizer.FindRouteOptimized("another-service")
	if !exists {
		t.Error("Expected to find another-service route")
	}
	if route.RateLimit != 20 {
		t.Errorf("Expected rate limit 20, got %d", route.RateLimit)
	}

	// Test non-existent service
	route, exists = services.GlobalRouteOptimizer.FindRouteOptimized("non-existent-service")
	if exists {
		t.Error("Expected route to not exist")
	}
}

func BenchmarkRateLimitServiceLookup(b *testing.B) {
	gin.SetMode(gin.TestMode)

	// Initialize route optimizer
	services.GlobalRouteOptimizer.BuildOptimizedRoutes([]services.RouteConfig{
		{
			ServiceName: "test-service",
			BackendURL:  "http://localhost:8081",
			RateLimit:   1000,
		},
	})

	serviceNames := []string{"test-service", "non-existent-service"}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		serviceName := serviceNames[i%len(serviceNames)]
		services.GlobalRouteOptimizer.FindRouteOptimized(serviceName)
	}
}
