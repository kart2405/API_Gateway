package main

import (
	"github.com/gin-gonic/gin"
	"github.com/kart2405/API_Gateway/internal/services"
)

func main() {
	r := gin.Default()

	// Reverse Proxy route
	r.Any("/proxy/:service/*proxyPath", services.ReverseProxyHandler)

	// Run Gateway on port 8080
	r.Run(":8080")
}
