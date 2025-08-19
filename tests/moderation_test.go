package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kart2405/API_Gateway/internal/middleware/moderation"
)

// TestModerationMiddlewareSetup initializes test environment for moderation tests
func TestModerationMiddlewareSetup(t *testing.T) {
	gin.SetMode(gin.TestMode)
}

// TestModerationMiddleware_SafeContent tests the middleware with safe content
func TestModerationMiddleware_SafeContent(t *testing.T) {
	TestModerationMiddlewareSetup(t)

	// Create router with moderation middleware
	r := gin.New()
	r.Use(moderation.ModerationMiddleware())

	// Add a test endpoint that echoes the request body
	r.POST("/test", func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to read body"})
			return
		}
		c.JSON(200, gin.H{
			"message": "Content processed successfully",
			"body":    string(body),
		})
	})

	testCases := []struct {
		name        string
		requestBody interface{}
		description string
	}{
		{
			name: "safe_json_content",
			requestBody: map[string]interface{}{
				"message": "Hello world!",
				"user":    "john_doe",
				"action":  "create_post",
			},
			description: "JSON content with safe text",
		},
		{
			name:        "safe_plain_text",
			requestBody: "This is a safe message with no harmful content.",
			description: "Plain text content that is safe",
		},
		{
			name: "safe_complex_json",
			requestBody: map[string]interface{}{
				"user": map[string]interface{}{
					"name":  "Alice",
					"email": "alice@example.com",
				},
				"post": map[string]interface{}{
					"title":   "My awesome blog post",
					"content": "This is a wonderful day to write about technology!",
					"tags":    []string{"tech", "blog", "programming"},
				},
			},
			description: "Complex nested JSON with safe content",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var requestBody io.Reader

			if str, ok := tc.requestBody.(string); ok {
				requestBody = strings.NewReader(str)
			} else {
				jsonData, err := json.Marshal(tc.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal test data: %v", err)
				}
				requestBody = bytes.NewBuffer(jsonData)
			}

			req, err := http.NewRequest("POST", "/test", requestBody)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200 for %s, got %d. Response: %s",
					tc.description, w.Code, w.Body.String())
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Errorf("Failed to parse response JSON: %v", err)
			}

			if response["message"] != "Content processed successfully" {
				t.Errorf("Expected success message for %s", tc.description)
			}
		})
	}
}

// TestModerationMiddleware_UnsafeContent tests the middleware with unsafe content
func TestModerationMiddleware_UnsafeContent(t *testing.T) {
	TestModerationMiddlewareSetup(t)

	// Create router with moderation middleware
	r := gin.New()
	r.Use(moderation.ModerationMiddleware())

	// Add a test endpoint (should not be reached for unsafe content)
	r.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "This should not be reached"})
	})

	testCases := []struct {
		name        string
		requestBody interface{}
		description string
	}{
		{
			name: "unsafe_json_spam",
			requestBody: map[string]interface{}{
				"message": "This is spam content that should be blocked",
				"user":    "spammer",
			},
			description: "JSON content with spam",
		},
		{
			name: "unsafe_json_hate",
			requestBody: map[string]interface{}{
				"comment": "I hate this so much",
				"rating":  1,
			},
			description: "JSON content with hate speech",
		},
		{
			name:        "unsafe_plain_text_violence",
			requestBody: "This message contains violence and should be blocked",
			description: "Plain text with violence",
		},
		{
			name: "unsafe_nested_json",
			requestBody: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "BadUser",
				},
				"posts": []interface{}{
					map[string]interface{}{
						"content": "This post contains spam content",
					},
				},
			},
			description: "Nested JSON with unsafe content",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var requestBody io.Reader

			if str, ok := tc.requestBody.(string); ok {
				requestBody = strings.NewReader(str)
			} else {
				jsonData, err := json.Marshal(tc.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal test data: %v", err)
				}
				requestBody = bytes.NewBuffer(jsonData)
			}

			req, err := http.NewRequest("POST", "/test", requestBody)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusForbidden {
				t.Errorf("Expected status 403 for %s, got %d. Response: %s",
					tc.description, w.Code, w.Body.String())
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Errorf("Failed to parse response JSON: %v", err)
			}

			expectedError := "Content violates our community guidelines"
			if response["error"] != expectedError {
				t.Errorf("Expected error message '%s' for %s, got '%v'",
					expectedError, tc.description, response["error"])
			}
		})
	}
}

// TestModerationMiddleware_EdgeCases tests edge cases
func TestModerationMiddleware_EdgeCases(t *testing.T) {
	TestModerationMiddlewareSetup(t)

	// Create router with moderation middleware
	r := gin.New()
	r.Use(moderation.ModerationMiddleware())

	// Add test endpoints for different HTTP methods
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "GET request processed"})
	})

	r.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "POST request processed"})
	})

	r.PUT("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "PUT request processed"})
	})

	t.Run("get_request_bypasses_moderation", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected GET request to bypass moderation, got status %d", w.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		if response["message"] != "GET request processed" {
			t.Error("GET request should bypass moderation middleware")
		}
	})

	t.Run("empty_body_post_request", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/test", strings.NewReader(""))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected empty body POST to pass through, got status %d", w.Code)
		}
	})

	t.Run("invalid_json_content", func(t *testing.T) {
		invalidJSON := `{"invalid": json content}`
		req, _ := http.NewRequest("POST", "/test", strings.NewReader(invalidJSON))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should be processed as plain text and checked for safety
		if w.Code != http.StatusOK {
			t.Errorf("Expected invalid JSON to be processed as text, got status %d", w.Code)
		}
	})

	t.Run("put_request_with_unsafe_content", func(t *testing.T) {
		unsafeContent := `{"message": "This contains spam content"}`
		req, _ := http.NewRequest("PUT", "/test", strings.NewReader(unsafeContent))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected PUT request with unsafe content to be blocked, got status %d", w.Code)
		}
	})
}

// TestModerationMiddleware_BodyPreservation tests that request body is preserved for downstream handlers
func TestModerationMiddleware_BodyPreservation(t *testing.T) {
	TestModerationMiddlewareSetup(t)

	// Create router with moderation middleware
	r := gin.New()
	r.Use(moderation.ModerationMiddleware())

	// Add endpoint that reads body twice to test preservation
	r.POST("/test", func(c *gin.Context) {
		// First read
		body1, err1 := io.ReadAll(c.Request.Body)
		if err1 != nil {
			t.Errorf("First body read failed: %v", err1)
		}

		// Reset body for second read (this is what the middleware should enable)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body1))

		// Second read
		body2, err2 := io.ReadAll(c.Request.Body)
		if err2 != nil {
			t.Errorf("Second body read failed: %v", err2)
		}

		if string(body1) != string(body2) {
			t.Errorf("Body preservation failed: first read '%s', second read '%s'",
				string(body1), string(body2))
		}

		c.JSON(200, gin.H{
			"message":    "Body preserved successfully",
			"body_size":  len(body1),
			"body_match": string(body1) == string(body2),
		})
	})

	testContent := map[string]interface{}{
		"message": "Test content for body preservation",
		"data":    []string{"item1", "item2", "item3"},
	}

	jsonData, _ := json.Marshal(testContent)
	req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["body_match"] != true {
		t.Error("Body preservation test failed - body content didn't match")
	}

	if response["body_size"].(float64) != float64(len(jsonData)) {
		t.Errorf("Body size mismatch: expected %d, got %f",
			len(jsonData), response["body_size"].(float64))
	}
}

// TestModerationMiddleware_Integration tests the middleware in a more realistic scenario
func TestModerationMiddleware_Integration(t *testing.T) {
	TestModerationMiddlewareSetup(t)

	// Create router with multiple middleware layers (simulating real usage)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(moderation.ModerationMiddleware())

	// Simulate a blog post creation endpoint
	r.POST("/api/posts", func(c *gin.Context) {
		var post struct {
			Title   string   `json:"title" binding:"required"`
			Content string   `json:"content" binding:"required"`
			Author  string   `json:"author" binding:"required"`
			Tags    []string `json:"tags"`
		}

		if err := c.ShouldBindJSON(&post); err != nil {
			c.JSON(400, gin.H{"error": "Invalid post data"})
			return
		}

		// Simulate post creation
		c.JSON(201, gin.H{
			"message": "Post created successfully",
			"post_id": 12345,
			"title":   post.Title,
			"author":  post.Author,
		})
	})

	// Simulate a comment creation endpoint
	r.POST("/api/comments", func(c *gin.Context) {
		var comment struct {
			PostID  int    `json:"post_id" binding:"required"`
			Content string `json:"content" binding:"required"`
			Author  string `json:"author" binding:"required"`
		}

		if err := c.ShouldBindJSON(&comment); err != nil {
			c.JSON(400, gin.H{"error": "Invalid comment data"})
			return
		}

		c.JSON(201, gin.H{
			"message":    "Comment created successfully",
			"comment_id": 67890,
			"post_id":    comment.PostID,
		})
	})

	testCases := []struct {
		name           string
		endpoint       string
		requestBody    interface{}
		expectedStatus int
		description    string
	}{
		{
			name:     "safe_blog_post",
			endpoint: "/api/posts",
			requestBody: map[string]interface{}{
				"title":   "Introduction to Go Programming",
				"content": "Go is a wonderful programming language developed by Google...",
				"author":  "john_doe",
				"tags":    []string{"golang", "programming", "tutorial"},
			},
			expectedStatus: 201,
			description:    "Safe blog post should be created successfully",
		},
		{
			name:     "unsafe_blog_post",
			endpoint: "/api/posts",
			requestBody: map[string]interface{}{
				"title":   "Spam Post",
				"content": "This is spam content that should be blocked",
				"author":  "spammer",
				"tags":    []string{"spam"},
			},
			expectedStatus: 403,
			description:    "Unsafe blog post should be blocked",
		},
		{
			name:     "safe_comment",
			endpoint: "/api/comments",
			requestBody: map[string]interface{}{
				"post_id": 123,
				"content": "Great article! Thanks for sharing this information.",
				"author":  "reader123",
			},
			expectedStatus: 201,
			description:    "Safe comment should be created successfully",
		},
		{
			name:     "unsafe_comment",
			endpoint: "/api/comments",
			requestBody: map[string]interface{}{
				"post_id": 123,
				"content": "I hate this article and it contains violence",
				"author":  "troll_user",
			},
			expectedStatus: 403,
			description:    "Unsafe comment should be blocked",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tc.requestBody)
			if err != nil {
				t.Fatalf("Failed to marshal test data: %v", err)
			}

			req, err := http.NewRequest("POST", tc.endpoint, bytes.NewBuffer(jsonData))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("%s: Expected status %d, got %d. Response: %s",
					tc.description, tc.expectedStatus, w.Code, w.Body.String())
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Errorf("Failed to parse response JSON: %v", err)
			}

			// Verify appropriate response messages
			if tc.expectedStatus == 403 {
				if response["error"] != "Content violates our community guidelines" {
					t.Errorf("Expected community guidelines error message")
				}
			} else if tc.expectedStatus == 201 {
				if !strings.Contains(response["message"].(string), "created successfully") {
					t.Errorf("Expected success message for safe content")
				}
			}
		})
	}
}

// BenchmarkModerationMiddleware benchmarks the moderation middleware performance
func BenchmarkModerationMiddleware(b *testing.B) {
	gin.SetMode(gin.TestMode)

	// Create router with moderation middleware
	r := gin.New()
	r.Use(moderation.ModerationMiddleware())
	r.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "OK"})
	})

	// Prepare test data
	testContent := map[string]interface{}{
		"message": "This is a safe message for benchmarking",
		"user":    "benchmark_user",
		"data":    []string{"item1", "item2", "item3"},
	}
	jsonData, _ := json.Marshal(testContent)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}
