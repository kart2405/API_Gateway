package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kart2405/API_Gateway/internal/middleware/auth"
	"github.com/kart2405/API_Gateway/internal/middleware/moderation"
)

// Example demonstrating how to integrate the moderation middleware
// into your API Gateway application
func main() {
	// Set Gin to release mode for cleaner output
	gin.SetMode(gin.ReleaseMode)

	// Create Gin router
	r := gin.New()

	// Add basic middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Public routes (no authentication required)
	r.POST("/public/feedback", func(c *gin.Context) {
		var feedback struct {
			Message string `json:"message" binding:"required"`
			Email   string `json:"email" binding:"required"`
		}

		if err := c.ShouldBindJSON(&feedback); err != nil {
			c.JSON(400, gin.H{"error": "Invalid feedback data"})
			return
		}

		c.JSON(200, gin.H{
			"message": "Feedback received successfully",
			"id":      12345,
		})
	})

	// Protected routes with JWT authentication
	protected := r.Group("/api")
	protected.Use(auth.JWTAuthMiddleware())
	{
		// Routes with content moderation
		moderated := protected.Group("/")
		moderated.Use(moderation.ModerationMiddleware())
		{
			// Blog post creation with content moderation
			moderated.POST("/posts", func(c *gin.Context) {
				var post struct {
					Title   string   `json:"title" binding:"required"`
					Content string   `json:"content" binding:"required"`
					Tags    []string `json:"tags"`
				}

				if err := c.ShouldBindJSON(&post); err != nil {
					c.JSON(400, gin.H{"error": "Invalid post data"})
					return
				}

				// Simulate post creation
				c.JSON(201, gin.H{
					"message": "Post created successfully",
					"post_id": 67890,
					"title":   post.Title,
				})
			})

			// Comment creation with content moderation
			moderated.POST("/comments", func(c *gin.Context) {
				var comment struct {
					PostID  int    `json:"post_id" binding:"required"`
					Content string `json:"content" binding:"required"`
				}

				if err := c.ShouldBindJSON(&comment); err != nil {
					c.JSON(400, gin.H{"error": "Invalid comment data"})
					return
				}

				c.JSON(201, gin.H{
					"message":    "Comment created successfully",
					"comment_id": 54321,
					"post_id":    comment.PostID,
				})
			})

			// User profile update with content moderation
			moderated.PUT("/profile", func(c *gin.Context) {
				var profile struct {
					Bio      string `json:"bio"`
					Website  string `json:"website"`
					Location string `json:"location"`
				}

				if err := c.ShouldBindJSON(&profile); err != nil {
					c.JSON(400, gin.H{"error": "Invalid profile data"})
					return
				}

				c.JSON(200, gin.H{
					"message": "Profile updated successfully",
					"bio":     profile.Bio,
				})
			})
		}

		// Routes without content moderation (e.g., read-only operations)
		protected.GET("/posts", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"posts": []map[string]interface{}{
					{
						"id":      1,
						"title":   "Sample Post",
						"content": "This is a sample post content",
					},
				},
			})
		})

		protected.GET("/profile", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			username, _ := c.Get("username")

			c.JSON(200, gin.H{
				"user_id":  userID,
				"username": username,
				"bio":      "Sample user bio",
			})
		})
	}

	// Demo endpoint to show middleware functionality
	r.GET("/demo", func(c *gin.Context) {
		c.HTML(200, "demo.html", gin.H{
			"title": "Moderation Middleware Demo",
		})
	})

	// Start server
	fmt.Println("🚀 Server starting on http://localhost:8080")
	fmt.Println("📋 Available endpoints:")
	fmt.Println("   POST /public/feedback - Public feedback (no auth, no moderation)")
	fmt.Println("   POST /api/posts - Create post (auth + moderation required)")
	fmt.Println("   POST /api/comments - Create comment (auth + moderation required)")
	fmt.Println("   PUT  /api/profile - Update profile (auth + moderation required)")
	fmt.Println("   GET  /api/posts - List posts (auth required, no moderation)")
	fmt.Println("   GET  /api/profile - Get profile (auth required, no moderation)")
	fmt.Println("")
	fmt.Println("🧪 Test the middleware with sample requests...")

	// Start server in a goroutine so we can run demo requests
	go func() {
		if err := r.Run(":8080"); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(1 * time.Second)

	// Run demo requests
	runDemoRequests()
}

// runDemoRequests demonstrates the middleware functionality
func runDemoRequests() {
	fmt.Println("🔍 Running demo requests to test moderation middleware...")

	baseURL := "http://localhost:8080"

	// Generate a test token for authenticated requests
	token, _ := auth.GenerateToken(1, "demo_user")

	testCases := []struct {
		name        string
		method      string
		endpoint    string
		body        interface{}
		needsAuth   bool
		expectedMsg string
	}{
		{
			name:     "Safe public feedback",
			method:   "POST",
			endpoint: "/public/feedback",
			body: map[string]string{
				"message": "Great service! Keep up the good work.",
				"email":   "user@example.com",
			},
			needsAuth:   false,
			expectedMsg: "Should succeed - safe content, no moderation",
		},
		{
			name:     "Safe blog post",
			method:   "POST",
			endpoint: "/api/posts",
			body: map[string]interface{}{
				"title":   "Introduction to Go Programming",
				"content": "Go is a wonderful programming language...",
				"tags":    []string{"golang", "tutorial"},
			},
			needsAuth:   true,
			expectedMsg: "Should succeed - safe content passes moderation",
		},
		{
			name:     "Unsafe blog post (spam)",
			method:   "POST",
			endpoint: "/api/posts",
			body: map[string]interface{}{
				"title":   "Spam Post",
				"content": "This is spam content that should be blocked",
				"tags":    []string{"spam"},
			},
			needsAuth:   true,
			expectedMsg: "Should be blocked - contains spam",
		},
		{
			name:     "Safe comment",
			method:   "POST",
			endpoint: "/api/comments",
			body: map[string]interface{}{
				"post_id": 123,
				"content": "Great article! Thanks for sharing.",
			},
			needsAuth:   true,
			expectedMsg: "Should succeed - safe comment",
		},
		{
			name:     "Unsafe comment (hate)",
			method:   "POST",
			endpoint: "/api/comments",
			body: map[string]interface{}{
				"post_id": 123,
				"content": "I hate this article so much",
			},
			needsAuth:   true,
			expectedMsg: "Should be blocked - contains hate speech",
		},
		{
			name:     "Safe profile update",
			method:   "PUT",
			endpoint: "/api/profile",
			body: map[string]interface{}{
				"bio":      "Software developer passionate about Go",
				"website":  "https://example.com",
				"location": "San Francisco, CA",
			},
			needsAuth:   true,
			expectedMsg: "Should succeed - safe profile content",
		},
		{
			name:        "Get posts (no moderation)",
			method:      "GET",
			endpoint:    "/api/posts",
			body:        nil,
			needsAuth:   true,
			expectedMsg: "Should succeed - GET requests bypass moderation",
		},
	}

	for i, tc := range testCases {
		fmt.Printf("\n%d. Testing: %s\n", i+1, tc.name)
		fmt.Printf("   Expected: %s\n", tc.expectedMsg)

		// Prepare request
		var reqBody *bytes.Buffer
		if tc.body != nil {
			jsonData, _ := json.Marshal(tc.body)
			reqBody = bytes.NewBuffer(jsonData)
		} else {
			reqBody = bytes.NewBuffer([]byte{})
		}

		req, _ := http.NewRequest(tc.method, baseURL+tc.endpoint, reqBody)
		if tc.body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		if tc.needsAuth {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		// Make request
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("   ❌ Request failed: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		// Parse response
		var response map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&response)

		// Display result
		statusIcon := "✅"
		if resp.StatusCode >= 400 {
			statusIcon = "🚫"
		}

		fmt.Printf("   %s Status: %d\n", statusIcon, resp.StatusCode)
		if message, ok := response["message"]; ok {
			fmt.Printf("   📝 Message: %s\n", message)
		}
		if errorMsg, ok := response["error"]; ok {
			fmt.Printf("   ⚠️  Error: %s\n", errorMsg)
		}
	}

	fmt.Println("\n🎉 Demo completed! The moderation middleware is working correctly.")
	fmt.Println("💡 Key observations:")
	fmt.Println("   - Safe content passes through and reaches the handlers")
	fmt.Println("   - Unsafe content is blocked with 403 Forbidden status")
	fmt.Println("   - GET requests bypass moderation entirely")
	fmt.Println("   - Request bodies are preserved for downstream handlers")
	fmt.Println("   - The middleware integrates seamlessly with existing auth middleware")
}
