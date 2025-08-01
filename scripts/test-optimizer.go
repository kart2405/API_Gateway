package main

import (
	"fmt"
	"log"

	"github.com/kart2405/API_Gateway/internal/config"
	"github.com/kart2405/API_Gateway/internal/services"
)

func main() {
	// Initialize database
	config.InitDatabase()
	config.DB.AutoMigrate(&services.RouteConfig{})

	// Load routes from database
	var routes []services.RouteConfig
	if err := config.DB.Find(&routes).Error; err != nil {
		log.Fatalf("Failed to load routes: %v", err)
	}

	fmt.Printf("Found %d routes in database:\n", len(routes))
	for _, route := range routes {
		fmt.Printf("- %s -> %s\n", route.ServiceName, route.BackendURL)
	}

	// Initialize route optimizer
	if err := services.InitializeRouteOptimizer(); err != nil {
		log.Fatalf("Failed to initialize route optimizer: %v", err)
	}

	// Check optimizer stats
	stats := services.GlobalRouteOptimizer.GetRouteStats()
	fmt.Printf("\nOptimizer stats: %+v\n", stats)

	// Test route lookup
	route, exists := services.GlobalRouteOptimizer.FindRouteOptimized("test-service")
	if exists {
		fmt.Printf("Found route: %s -> %s\n", route.ServiceName, route.BackendURL)
	} else {
		fmt.Println("Route not found in optimizer")
	}
}
