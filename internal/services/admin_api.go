package services

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kart2405/API_Gateway/internal/config"
)

type RouteConfig struct {
	ID              uint   `json:"id" gorm:"primaryKey"`
	ServiceName     string `json:"service_name" gorm:"uniqueIndex;not null"`
	BackendURL      string `json:"backend_url" gorm:"not null"`
	RateLimit       int    `json:"rate_limit" gorm:"default:10"`
	RateLimitWindow int    `json:"rate_limit_window" gorm:"default:60"`
	IsActive        bool   `json:"is_active" gorm:"default:true"`
}

// AdminGetRoutes returns all configured routes
func AdminGetRoutes(c *gin.Context) {
	var routes []RouteConfig
	if err := config.DB.Find(&routes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch routes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"routes": routes,
		"count":  len(routes),
	})
}

// AdminCreateRoute creates a new route configuration
func AdminCreateRoute(c *gin.Context) {
	var route RouteConfig
	if err := c.ShouldBindJSON(&route); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Validate required fields
	if route.ServiceName == "" || route.BackendURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Service name and backend URL are required"})
		return
	}

	// Check if route already exists
	var existingRoute RouteConfig
	if err := config.DB.Where("service_name = ?", route.ServiceName).First(&existingRoute).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Route with this service name already exists"})
		return
	}

	// Set defaults
	if route.RateLimit == 0 {
		route.RateLimit = 10
	}
	if route.RateLimitWindow == 0 {
		route.RateLimitWindow = 60
	}

	// Create route
	if err := config.DB.Create(&route).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create route"})
		return
	}

	// Update in-memory route map
	config.RouteMap[route.ServiceName] = route.BackendURL

	// Reload all routes in the optimizer
	var allRoutes []RouteConfig
	if err := config.DB.Find(&allRoutes).Error; err == nil {
		GlobalRouteOptimizer.BuildOptimizedRoutes(allRoutes)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Route created successfully",
		"route":   route,
	})
}

// AdminUpdateRoute updates an existing route configuration
func AdminUpdateRoute(c *gin.Context) {
	id := c.Param("id")
	routeID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid route ID"})
		return
	}

	var route RouteConfig
	if err := config.DB.First(&route, routeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Route not found"})
		return
	}

	var updateData RouteConfig
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Update fields
	if updateData.ServiceName != "" {
		route.ServiceName = updateData.ServiceName
	}
	if updateData.BackendURL != "" {
		route.BackendURL = updateData.BackendURL
	}
	if updateData.RateLimit > 0 {
		route.RateLimit = updateData.RateLimit
	}
	if updateData.RateLimitWindow > 0 {
		route.RateLimitWindow = updateData.RateLimitWindow
	}

	// Save changes
	if err := config.DB.Save(&route).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update route"})
		return
	}

	// Update in-memory route map
	config.RouteMap[route.ServiceName] = route.BackendURL

	// Reload all routes in the optimizer
	var allRoutes []RouteConfig
	if err := config.DB.Find(&allRoutes).Error; err == nil {
		GlobalRouteOptimizer.BuildOptimizedRoutes(allRoutes)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Route updated successfully",
		"route":   route,
	})
}

// AdminDeleteRoute deletes a route configuration
func AdminDeleteRoute(c *gin.Context) {
	id := c.Param("id")
	routeID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid route ID"})
		return
	}

	var route RouteConfig
	if err := config.DB.First(&route, routeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Route not found"})
		return
	}

	// Delete route
	if err := config.DB.Delete(&route).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete route"})
		return
	}

	// Remove from in-memory route map
	delete(config.RouteMap, route.ServiceName)

	// Reload all routes in the optimizer
	var allRoutes []RouteConfig
	if err := config.DB.Find(&allRoutes).Error; err == nil {
		GlobalRouteOptimizer.BuildOptimizedRoutes(allRoutes)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Route deleted successfully",
	})
}

// AdminGetRouteStats returns statistics about route usage
func AdminGetRouteStats(c *gin.Context) {
	var routeCount int64
	config.DB.Model(&RouteConfig{}).Count(&routeCount)

	var activeRoutes int64
	config.DB.Model(&RouteConfig{}).Where("is_active = ?", true).Count(&activeRoutes)

	c.JSON(http.StatusOK, gin.H{
		"total_routes":    routeCount,
		"active_routes":   activeRoutes,
		"inactive_routes": routeCount - activeRoutes,
	})
}
