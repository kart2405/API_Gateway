package services

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func ReverseProxyHandler(c *gin.Context) {
	serviceName := c.Param("service") // e.g., service1
	proxyPath := c.Param("proxyPath") // e.g., /users/123

	// Use optimized route lookup instead of simple map lookup
	route, exists := GlobalRouteOptimizer.FindRouteOptimized(serviceName)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found in config"})
		return
	}

	// Use the route configuration for backend URL
	backendBaseURL := route.BackendURL

	// Construct full URL to forward to
	fullBackendURL := backendBaseURL + proxyPath
	fmt.Println("Forwarding to:", fullBackendURL)

	// Create the proxy request
	req, err := http.NewRequest(c.Request.Method, fullBackendURL, c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	// Copy headers from original request
	for k, v := range c.Request.Header {
		req.Header[k] = v
	}

	// Send request to backend
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to reach backend"})
		return
	}
	defer resp.Body.Close()

	// Copy status, headers, and body back to client
	c.Status(resp.StatusCode)
	for k, v := range resp.Header {
		c.Header(k, strings.Join(v, ","))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read backend response"})
		return
	}
	c.Writer.Write(body)

	fmt.Println("Responded with status:", resp.StatusCode)
}
