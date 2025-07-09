package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kart2405/API_Gateway/internal/middleware/auth"
	"github.com/kart2405/API_Gateway/internal/services"
)

// TestSetup initializes test environment without database
func TestSetup(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Use a real route optimizer but with mock data
	services.GlobalRouteOptimizer = services.NewRouteOptimizer()
}

// TestLoginEndpoint tests the login functionality with mock data
func TestLoginEndpoint(t *testing.T) {
	TestSetup(t)

	// Create router with mock login handler
	r := gin.New()
	r.POST("/login", func(c *gin.Context) {
		var login struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&login); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}

		// Mock authentication
		if login.Username == "testuser" && login.Password == "testpass" {
			token, _ := auth.GenerateToken(1, "testuser")
			c.JSON(200, gin.H{
				"token":    token,
				"userID":   1,
				"username": "testuser",
			})
		} else {
			c.JSON(401, gin.H{"error": "Invalid username or password"})
		}
	})

	// Test valid login
	loginData := map[string]string{
		"username": "testuser",
		"password": "testpass",
	}
	jsonData, _ := json.Marshal(loginData)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["token"] == "" {
		t.Error("Expected token in response")
	}

	// Test invalid login
	invalidLoginData := map[string]string{
		"username": "testuser",
		"password": "wrongpass",
	}
	jsonData, _ = json.Marshal(invalidLoginData)

	req, _ = http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

// TestJWTAuthMiddleware tests JWT authentication middleware
func TestJWTAuthMiddleware(t *testing.T) {
	TestSetup(t)

	// Create test token
	token, _ := auth.GenerateToken(1, "testuser")

	// Create router with protected endpoint
	r := gin.New()
	protected := r.Group("/")
	protected.Use(auth.JWTAuthMiddleware())
	{
		protected.GET("/protected", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			username, _ := c.Get("username")
			c.JSON(200, gin.H{
				"message":  "Protected endpoint accessed",
				"userID":   userID,
				"username": username,
			})
		})
	}

	// Test with valid token
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test without token
	req, _ = http.NewRequest("GET", "/protected", nil)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	// Test with invalid token
	req, _ = http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

// TestRouteOptimizer tests the route optimization functionality
func TestRouteOptimizer(t *testing.T) {
	TestSetup(t)

	// Create test routes
	routes := []services.RouteConfig{
		{
			ServiceName: "user-service",
			BackendURL:  "http://localhost:8081",
			RateLimit:   10,
		},
		{
			ServiceName: "order-service",
			BackendURL:  "http://localhost:8082",
			RateLimit:   15,
		},
		{
			ServiceName: "payment-service",
			BackendURL:  "http://localhost:8083",
			RateLimit:   20,
		},
	}

	// Build optimized routes
	optimizer := services.NewRouteOptimizer()
	optimizer.BuildOptimizedRoutes(routes)

	// Test hash-map lookup
	route, exists := optimizer.FindRouteByHashMap("user-service")
	if !exists {
		t.Error("Expected to find user-service route")
	}
	if route.BackendURL != "http://localhost:8081" {
		t.Error("Expected correct backend URL")
	}

	// Test optimized lookup
	route, exists = optimizer.FindRouteOptimized("order-service")
	if !exists {
		t.Error("Expected to find order-service route")
	}
	if route.BackendURL != "http://localhost:8082" {
		t.Error("Expected correct backend URL")
	}

	// Test non-existent route
	route, exists = optimizer.FindRouteOptimized("non-existent-service")
	if exists {
		t.Error("Expected route to not exist")
	}

	// Test benchmark functionality
	serviceNames := []string{"user-service", "order-service", "payment-service"}
	benchmark := optimizer.BenchmarkRouteLookup(serviceNames)

	if benchmark["improvement_percentage"] == 0 {
		t.Error("Expected improvement percentage to be calculated")
	}
}

// TestAdminAPIs tests admin API endpoints with mock data
func TestAdminAPIs(t *testing.T) {
	TestSetup(t)

	// Create test token
	token, _ := auth.GenerateToken(1, "admin")

	// Create router with admin endpoints
	r := gin.New()
	admin := r.Group("/admin")
	admin.Use(auth.JWTAuthMiddleware())
	{
		admin.GET("/routes", func(c *gin.Context) {
			routes := []services.RouteConfig{
				{
					ID:          1,
					ServiceName: "test-service",
					BackendURL:  "http://localhost:8081",
					RateLimit:   20,
				},
			}
			c.JSON(200, gin.H{
				"routes": routes,
				"count":  len(routes),
			})
		})

		admin.POST("/routes", func(c *gin.Context) {
			var route services.RouteConfig
			if err := c.ShouldBindJSON(&route); err != nil {
				c.JSON(400, gin.H{"error": "Invalid request data"})
				return
			}

			if route.ServiceName == "" || route.BackendURL == "" {
				c.JSON(400, gin.H{"error": "Service name and backend URL are required"})
				return
			}

			c.JSON(201, gin.H{
				"message": "Route created successfully",
				"route":   route,
			})
		})

		admin.GET("/routes/stats", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"total_routes":    1,
				"active_routes":   1,
				"inactive_routes": 0,
			})
		})
	}

	// Test creating a route
	routeData := map[string]interface{}{
		"service_name": "test-service",
		"backend_url":  "http://localhost:8081",
		"rate_limit":   20,
	}
	jsonData, _ := json.Marshal(routeData)

	req, _ := http.NewRequest("POST", "/admin/routes", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	// Test getting routes
	req, _ = http.NewRequest("GET", "/admin/routes", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["count"].(float64) < 1 {
		t.Error("Expected at least one route")
	}

	// Test getting stats
	req, _ = http.NewRequest("GET", "/admin/routes/stats", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestReverseProxy tests the reverse proxy functionality
func TestReverseProxy(t *testing.T) {
	TestSetup(t)

	// Create test token
	token, _ := auth.GenerateToken(1, "testuser")

	// Set up route optimizer with test route
	services.GlobalRouteOptimizer.BuildOptimizedRoutes([]services.RouteConfig{
		{
			ServiceName: "test-service",
			BackendURL:  "http://localhost:8081",
			RateLimit:   10,
		},
	})

	// Create router with reverse proxy
	r := gin.New()
	protected := r.Group("/")
	protected.Use(auth.JWTAuthMiddleware())
	{
		protected.Any("/proxy/:service/*proxyPath", services.ReverseProxyHandler)
	}

	// Test reverse proxy (this would require a mock backend server)
	req, _ := http.NewRequest("GET", "/proxy/test-service/api/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Since we don't have a real backend, this should fail with 502
	// In a real test, you'd set up a mock backend server
	if w.Code != http.StatusBadGateway {
		t.Errorf("Expected status 502 (no backend), got %d", w.Code)
	}
}

// BenchmarkRouteLookup benchmarks the route lookup performance
func BenchmarkRouteLookup(b *testing.B) {
	TestSetup(&testing.T{})

	// Create test routes
	routes := []services.RouteConfig{
		{ServiceName: "user-service", BackendURL: "http://localhost:8081"},
		{ServiceName: "order-service", BackendURL: "http://localhost:8082"},
		{ServiceName: "payment-service", BackendURL: "http://localhost:8083"},
		{ServiceName: "inventory-service", BackendURL: "http://localhost:8084"},
		{ServiceName: "notification-service", BackendURL: "http://localhost:8085"},
	}

	optimizer := services.NewRouteOptimizer()
	optimizer.BuildOptimizedRoutes(routes)

	serviceNames := []string{"user-service", "order-service", "payment-service", "inventory-service", "notification-service"}

	b.ResetTimer()

	// Benchmark optimized lookup
	b.Run("Optimized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			serviceName := serviceNames[i%len(serviceNames)]
			optimizer.FindRouteOptimized(serviceName)
		}
	})

	// Benchmark hash-map lookup
	b.Run("HashMap", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			serviceName := serviceNames[i%len(serviceNames)]
			optimizer.FindRouteByHashMap(serviceName)
		}
	})

	// Benchmark prefix tree lookup
	b.Run("PrefixTree", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			serviceName := serviceNames[i%len(serviceNames)]
			optimizer.FindRouteByPrefixTree(serviceName)
		}
	})
}
