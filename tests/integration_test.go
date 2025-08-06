package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kart2405/API_Gateway/internal/config"
	"github.com/kart2405/API_Gateway/internal/middleware/cache"
	"github.com/kart2405/API_Gateway/internal/middleware/health"
	"github.com/kart2405/API_Gateway/internal/middleware/logging"
	"github.com/kart2405/API_Gateway/internal/middleware/validation"
	"github.com/kart2405/API_Gateway/internal/services"
	"go.uber.org/zap"
)

// TestIntegrationSetup sets up the integration test environment
func TestIntegrationSetup(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Load configuration
	if err := config.LoadConfig(); err != nil {
		t.Skipf("Skipping integration test - config not loaded: %v", err)
	}

	// Initialize logger
	if err := logging.InitLogger(); err != nil {
		t.Skipf("Skipping integration test - logger not initialized: %v", err)
	}

	// Initialize database (skip if not available)
	config.InitDatabase()

	// Initialize Redis (skip if not available)
	config.InitRedis()

	// Initialize route optimizer
	services.GlobalRouteOptimizer.BuildOptimizedRoutes([]services.RouteConfig{
		{
			ServiceName: "test-service",
			BackendURL:  "http://localhost:8081",
			RateLimit:   10,
		},
	})
}

// TestHealthCheckSystem tests the health check system
func TestHealthCheckSystem(t *testing.T) {
	TestIntegrationSetup(t)

	// Create health service
	healthService := health.NewHealthService()

	// Create router with health checks
	r := gin.New()
	r.Use(health.HealthMiddleware(healthService))

	// Test health endpoint
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should return 200 or 503 depending on service availability
	if w.Code != http.StatusOK && w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 200 or 503, got %d", w.Code)
	}

	// Test readiness endpoint
	req, _ = http.NewRequest("GET", "/health/ready", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 200 or 503, got %d", w.Code)
	}

	// Test liveness endpoint
	req, _ = http.NewRequest("GET", "/health/live", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestValidationSystem tests the validation system
func TestValidationSystem(t *testing.T) {
	TestIntegrationSetup(t)

	// Create validation service
	validationService := validation.CreateCommonValidationService()

	// Create router with validation
	r := gin.New()
	r.Use(validation.ValidationMiddleware(validationService))

	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// Test valid request
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test validation with invalid email
	req, _ = http.NewRequest("GET", "/test?email=invalid-email", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should fail validation since email is invalid
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid email, got %d", w.Code)
	}
}

// TestCacheSystem tests the caching system
func TestCacheSystem(t *testing.T) {
	TestIntegrationSetup(t)

	// Create cache service
	cacheService := cache.NewCacheService()

	// Create router with caching
	r := gin.New()
	r.Use(cache.CacheMiddleware(cacheService, 5*time.Minute))

	r.GET("/cache-test", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "cached response",
			"timestamp": time.Now().Unix(),
		})
	})

	// First request - should not be cached
	req, _ := http.NewRequest("GET", "/cache-test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Second request - should be cached
	req, _ = http.NewRequest("GET", "/cache-test", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test cache stats
	stats := cacheService.GetStats()
	if stats["enabled"] != true {
		t.Error("Cache should be enabled")
	}
}

// TestLoggingSystem tests the logging system
func TestLoggingSystem(t *testing.T) {
	TestIntegrationSetup(t)

	// Create router with logging
	r := gin.New()
	r.Use(logging.RequestLogger())

	r.GET("/log-test", func(c *gin.Context) {
		logger := logging.GetLoggerFromContext(c)
		logger.Info("Test log message",
			zap.String("test", "integration"),
			zap.Int("status", 200),
		)
		c.JSON(200, gin.H{"message": "logged"})
	})

	// Test request logging
	req, _ := http.NewRequest("GET", "/log-test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check if request ID was added
	if w.Header().Get("X-Request-ID") == "" {
		t.Error("Expected X-Request-ID header")
	}
}

// TestConfigurationSystem tests the configuration system
func TestConfigurationSystem(t *testing.T) {
	TestIntegrationSetup(t)

	// Test configuration loading
	appConfig := config.GetConfig()
	if appConfig == nil {
		t.Fatal("Configuration should be loaded")
	}

	// Test server configuration
	if appConfig.Server.Port == "" {
		t.Error("Server port should be configured")
	}

	// Test database configuration
	if appConfig.Database.Host == "" {
		t.Error("Database host should be configured")
	}

	// Test Redis configuration
	if appConfig.Redis.Host == "" {
		t.Error("Redis host should be configured")
	}

	// Test security configuration
	if appConfig.Security.JWTSecret == "" {
		t.Error("JWT secret should be configured")
	}

	// Test logging configuration
	if appConfig.Logging.Level == "" {
		t.Error("Logging level should be configured")
	}
}

// TestRequestValidation tests request validation
func TestRequestValidation(t *testing.T) {
	TestIntegrationSetup(t)

	// Create validation service
	validationService := validation.NewValidationService()
	validationService.AddRule("email", &validation.EmailValidator{})
	validationService.AddRule("username", &validation.RequiredValidator{}, &validation.LengthValidator{Min: 3, Max: 50})

	// Create router with body validation
	r := gin.New()
	r.Use(validation.BodyValidationMiddleware(validationService))

	r.POST("/validate", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "validated"})
	})

	// Test valid request
	validData := map[string]string{
		"email":    "test@example.com",
		"username": "testuser",
	}
	jsonData, _ := json.Marshal(validData)

	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test invalid request
	invalidData := map[string]string{
		"email":    "invalid-email",
		"username": "ab", // Too short
	}
	jsonData, _ = json.Marshal(invalidData)

	req, _ = http.NewRequest("POST", "/validate", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should fail validation
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// TestSanitization tests input sanitization
func TestSanitization(t *testing.T) {
	// Test string sanitization
	testCases := []struct {
		input    string
		expected string
	}{
		{"  test  ", "test"},
		{"test\x00string", "teststring"},
		{"test\nstring", "test\nstring"}, // Newline should be preserved
		{"test\x01string", "teststring"}, // Control character should be removed
	}

	for _, tc := range testCases {
		result := validation.SanitizeString(tc.input)
		if result != tc.expected {
			t.Errorf("SanitizeString(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}

	// Test map sanitization
	inputMap := map[string]string{
		"key1": "  value1  ",
		"key2": "value2\x00",
		"key3": "value3",
	}

	sanitizedMap := validation.SanitizeMap(inputMap)
	expectedMap := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	for key, expected := range expectedMap {
		if sanitizedMap[key] != expected {
			t.Errorf("SanitizeMap: key %s = %q, expected %q", key, sanitizedMap[key], expected)
		}
	}
}

// BenchmarkCachePerformance benchmarks cache performance
func BenchmarkCachePerformance(b *testing.B) {
	TestIntegrationSetup(&testing.T{})

	cacheService := cache.NewCacheService()
	ctx := context.Background()

	b.ResetTimer()

	b.Run("Set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("benchmark_key_%d", i)
			cacheService.Set(ctx, key, "test_value", 5*time.Minute)
		}
	})

	b.Run("Get", func(b *testing.B) {
		// Pre-populate cache
		for i := 0; i < 1000; i++ {
			key := fmt.Sprintf("benchmark_key_%d", i)
			cacheService.Set(ctx, key, "test_value", 5*time.Minute)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("benchmark_key_%d", i%1000)
			cacheService.Get(ctx, key)
		}
	})
}

// BenchmarkValidationPerformance benchmarks validation performance
func BenchmarkValidationPerformance(b *testing.B) {
	TestIntegrationSetup(&testing.T{})

	validationService := validation.CreateCommonValidationService()

	testData := map[string]string{
		"email":    "test@example.com",
		"username": "testuser123",
		"password": "securepassword123",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		validationService.Validate(testData)
	}
} 