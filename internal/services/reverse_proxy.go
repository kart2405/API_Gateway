package services

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ReverseProxyHandler(c *gin.Context) {
	// Define your backend URL (for testing you can hardcode one backend for now)
	backendURL := "http://localhost:8001" + c.Request.RequestURI[len("/proxy"):]
	// fmt.Println(backendURL)
	// Create HTTP client
	client := &http.Client{}

	// Create new request → same method, same body
	req, err := http.NewRequest(c.Request.Method, backendURL, c.Request.Body)
	// This makes a new request:

	// c.Request.Method → same method as original request → GET, POST, PUT...
	// backendURL → the URL we just built.
	// c.Request.Body → same request body → so if client sends POST data → we send it to the backend.
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Copy original headers
	req.Header = c.Request.Header

	// Send request to backend
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	// Read backend response
	body, _ := io.ReadAll(resp.Body)

	// Send backend response back to client
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}
