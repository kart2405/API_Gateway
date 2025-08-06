package health

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kart2405/API_Gateway/internal/config"
	"github.com/kart2405/API_Gateway/internal/middleware/logging"
	"github.com/kart2405/API_Gateway/internal/services"
	"go.uber.org/zap"
)

// HealthStatus represents the health status of a component
type HealthStatus struct {
	Status    string            `json:"status"`
	Message   string            `json:"message,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
	Details   map[string]string `json:"details,omitempty"`
}

// HealthChecker interface for health checks
type HealthChecker interface {
	Check(ctx context.Context) error
	Name() string
}

// DatabaseHealthChecker checks database health
type DatabaseHealthChecker struct{}

func (d *DatabaseHealthChecker) Name() string {
	return "database"
}

func (d *DatabaseHealthChecker) Check(ctx context.Context) error {
	if config.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	sqlDB, err := config.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	return sqlDB.PingContext(ctx)
}

// RedisHealthChecker checks Redis health
type RedisHealthChecker struct{}

func (r *RedisHealthChecker) Name() string {
	return "redis"
}

func (r *RedisHealthChecker) Check(ctx context.Context) error {
	if config.RedisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	return config.HealthCheckRedis()
}

// RouteOptimizerHealthChecker checks route optimizer health
type RouteOptimizerHealthChecker struct{}

func (ro *RouteOptimizerHealthChecker) Name() string {
	return "route_optimizer"
}

func (ro *RouteOptimizerHealthChecker) Check(ctx context.Context) error {
	stats := services.GlobalRouteOptimizer.GetRouteStats()
	if stats == nil {
		return fmt.Errorf("route optimizer not initialized")
	}
	return nil
}

// HealthService manages health checks
type HealthService struct {
	checkers []HealthChecker
	mutex    sync.RWMutex
}

// NewHealthService creates a new health service
func NewHealthService() *HealthService {
	return &HealthService{
		checkers: []HealthChecker{
			&DatabaseHealthChecker{},
			&RedisHealthChecker{},
			&RouteOptimizerHealthChecker{},
		},
	}
}

// AddChecker adds a new health checker
func (hs *HealthService) AddChecker(checker HealthChecker) {
	hs.mutex.Lock()
	defer hs.mutex.Unlock()
	hs.checkers = append(hs.checkers, checker)
}

// CheckAll performs health checks on all components
func (hs *HealthService) CheckAll(ctx context.Context) map[string]HealthStatus {
	hs.mutex.RLock()
	checkers := make([]HealthChecker, len(hs.checkers))
	copy(checkers, hs.checkers)
	hs.mutex.RUnlock()

	results := make(map[string]HealthStatus)
	var wg sync.WaitGroup

	for _, checker := range checkers {
		wg.Add(1)
		go func(c HealthChecker) {
			defer wg.Done()
			status := hs.checkComponent(ctx, c)
			results[c.Name()] = status
		}(checker)
	}

	wg.Wait()
	return results
}

// checkComponent checks a single component
func (hs *HealthService) checkComponent(ctx context.Context, checker HealthChecker) HealthStatus {
	start := time.Now()
	err := checker.Check(ctx)
	duration := time.Since(start)

	status := HealthStatus{
		Timestamp: time.Now(),
		Details: map[string]string{
			"duration": duration.String(),
		},
	}

	if err != nil {
		status.Status = "unhealthy"
		status.Message = err.Error()
	} else {
		status.Status = "healthy"
		status.Message = "Component is functioning normally"
	}

	return status
}

// IsHealthy checks if all components are healthy
func (hs *HealthService) IsHealthy(ctx context.Context) bool {
	results := hs.CheckAll(ctx)
	for _, status := range results {
		if status.Status != "healthy" {
			return false
		}
	}
	return true
}

// HealthMiddleware provides health check endpoints
func HealthMiddleware(healthService *HealthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		switch path {
		case "/health":
			handleHealthCheck(c, healthService)
		case "/health/ready":
			handleReadinessCheck(c, healthService)
		case "/health/live":
			handleLivenessCheck(c)
		default:
			c.Next()
		}
	}
}

// handleHealthCheck handles the main health check endpoint
func handleHealthCheck(c *gin.Context, healthService *HealthService) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	results := healthService.CheckAll(ctx)
	overallHealthy := healthService.IsHealthy(ctx)

	response := gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
		"components": results,
	}

	if !overallHealthy {
		response["status"] = "unhealthy"
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	c.JSON(http.StatusOK, response)
}

// handleReadinessCheck handles the readiness check endpoint
func handleReadinessCheck(c *gin.Context, healthService *HealthService) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if healthService.IsHealthy(ctx) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ready",
			"timestamp": time.Now(),
		})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "not_ready",
			"timestamp": time.Now(),
		})
	}
}

// handleLivenessCheck handles the liveness check endpoint
func handleLivenessCheck(c *gin.Context) {
	// Liveness check is simple - if we can respond, we're alive
	c.JSON(http.StatusOK, gin.H{
		"status":    "alive",
		"timestamp": time.Now(),
	})
}

// LogHealthStatus logs health check results
func LogHealthStatus(results map[string]HealthStatus) {
	logger := logging.GetLogger()
	
	for component, status := range results {
		fields := []zap.Field{
			zap.String("component", component),
			zap.String("status", status.Status),
			zap.Time("timestamp", status.Timestamp),
		}

		if status.Message != "" {
			fields = append(fields, zap.String("message", status.Message))
		}

		if status.Status == "healthy" {
			logger.Info("Health check passed", fields...)
		} else {
			logger.Error("Health check failed", fields...)
		}
	}
} 