package services

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Mapping server links for dynamic routing
var serviceRoutes = map[string]string{
	"service1": "http://localhost:8001",
	"service2": "http://localhost:8002",
	"service3": "http://localhost:8003",
}

func ReverseProxyHandler(c *gin.Context) {
	serviceName := c.Param("service")
	proxyPath := c.Param("proxyPath")

	backendBaseURL, ok := serviceRoutes[serviceName]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	fullBackendURL := backendBaseURL + proxyPath
	fmt.Println("Forwarding to:", fullBackendURL)

	// Create the new request
	req, err := http.NewRequest(c.Request.Method, fullBackendURL, c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Request creation failed"})
		return
	}

	// Copy headers
	for k, v := range c.Request.Header {
		req.Header[k] = v
	}

	// Forward the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Backend unreachable"})
		return
	}
	defer resp.Body.Close()

	// Copy response status and body
	c.Status(resp.StatusCode)
	for k, v := range resp.Header {
		c.Header(k, strings.Join(v, ","))
	}
	body, _ := io.ReadAll(resp.Body)
	c.Writer.Write(body)

	fmt.Println("Backend responded with:", resp.StatusCode)
}
