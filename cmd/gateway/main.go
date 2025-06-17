package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/kart2405/API_Gateway/internal/config"
	"github.com/kart2405/API_Gateway/internal/services"
)

func main() {
	// Load route map from YAML config
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create Gin router
	r := gin.Default()

	// Reverse Proxy route (uses dynamically loaded config.RouteMap)
	r.Any("/proxy/:service/*proxyPath", services.ReverseProxyHandler)

	// Start the gateway server on port 8080
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
