package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kart2405/API_Gateway/internal/middleware/moderation"
)

// Mock LLM server for testing
func createMockLLMServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse request to determine response
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		// Extract content from different request formats
		content := ""
		if messages, ok := body["messages"].([]interface{}); ok {
			// OpenAI format
			for _, msg := range messages {
				if msgMap, ok := msg.(map[string]interface{}); ok {
					if msgMap["role"] == "user" {
						content = msgMap["content"].(string)
					}
				}
			}
		} else if inputs, ok := body["inputs"].(string); ok {
			// Hugging Face format
			content = inputs
		} else if prompt, ok := body["prompt"].(string); ok {
			// Local model format
			content = prompt
		}

		// Mock response based on content
		isSafe := !strings.Contains(strings.ToLower(content), "spam") &&
			!strings.Contains(strings.ToLower(content), "hate") &&
			!strings.Contains(strings.ToLower(content), "violence")

		confidence := 0.95
		if !isSafe {
			confidence = 0.85
		}

		// Mock different response formats based on endpoint
		if strings.Contains(r.URL.Path, "chat/completions") {
			// OpenAI format
			response := map[string]interface{}{
				"choices": []map[string]interface{}{
					{
						"message": map[string]interface{}{
							"content": fmt.Sprintf(`{
								"is_safe": %t,
								"confidence": %.2f,
								"categories": {
									"spam": %.2f,
									"hate_speech": %.2f,
									"violence": %.2f,
									"harassment": 0.1,
									"adult_content": 0.05,
									"misinformation": 0.1
								},
								"reasoning": "Content analysis completed",
								"suggestions": "Consider rephrasing for better clarity"
							}`, isSafe, confidence,
								map[bool]float64{true: 0.1, false: 0.9}[isSafe],
								map[bool]float64{true: 0.05, false: 0.8}[isSafe],
								map[bool]float64{true: 0.1, false: 0.85}[isSafe]),
						},
					},
				},
			}
			json.NewEncoder(w).Encode(response)
		} else if strings.Contains(r.URL.Path, "messages") {
			// Anthropic format
			response := map[string]interface{}{
				"content": []map[string]interface{}{
					{
						"text": fmt.Sprintf(`{
							"is_safe": %t,
							"confidence": %.2f,
							"categories": {"spam": 0.1},
							"reasoning": "Claude analysis completed"
						}`, isSafe, confidence),
					},
				},
			}
			json.NewEncoder(w).Encode(response)
		} else {
			// Hugging Face or Local format
			response := []map[string]interface{}{
				{
					"generated_text": fmt.Sprintf(`{
						"is_safe": %t,
						"confidence": %.2f,
						"categories": {"spam": 0.1},
						"reasoning": "Local model analysis completed"
					}`, isSafe, confidence),
				},
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
}

// Test LLM moderation middleware with different providers
func TestLLMModerationMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Start mock LLM server
	mockServer := createMockLLMServer()
	defer mockServer.Close()

	testConfigs := map[string]moderation.LLMConfig{
		"openai": {
			Provider:    "openai",
			APIKey:      "test-key",
			BaseURL:     mockServer.URL,
			Model:       "gpt-4",
			MaxTokens:   500,
			Temperature: 0.1,
			Timeout:     5 * time.Second,
		},
		"claude": {
			Provider:    "anthropic",
			APIKey:      "test-key",
			BaseURL:     mockServer.URL,
			Model:       "claude-3-sonnet",
			MaxTokens:   500,
			Temperature: 0.1,
			Timeout:     5 * time.Second,
		},
		"huggingface": {
			Provider:    "huggingface",
			APIKey:      "test-key",
			BaseURL:     mockServer.URL,
			Model:       "test-model",
			MaxTokens:   500,
			Temperature: 0.1,
			Timeout:     5 * time.Second,
		},
	}

	for providerName, config := range testConfigs {
		t.Run(providerName, func(t *testing.T) {
			// Create router with LLM middleware
			r := gin.New()
			r.Use(moderation.LLMModerationMiddleware(config))

			r.POST("/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Content approved by LLM"})
			})

			// Test safe content
			t.Run("safe_content", func(t *testing.T) {
				safeContent := map[string]interface{}{
					"message": "This is a great article about programming!",
					"author":  "john_doe",
				}

				jsonData, _ := json.Marshal(safeContent)
				req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)

				if w.Code != http.StatusOK {
					t.Errorf("Expected 200 for safe content with %s, got %d. Response: %s",
						providerName, w.Code, w.Body.String())
				}
			})

			// Test unsafe content
			t.Run("unsafe_content", func(t *testing.T) {
				unsafeContent := map[string]interface{}{
					"message": "This is spam content that should be blocked",
					"author":  "spammer",
				}

				jsonData, _ := json.Marshal(unsafeContent)
				req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)

				if w.Code != http.StatusForbidden {
					t.Errorf("Expected 403 for unsafe content with %s, got %d. Response: %s",
						providerName, w.Code, w.Body.String())
				}

				// Check response includes LLM reasoning
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)

				if _, hasReasoning := response["reasoning"]; !hasReasoning {
					t.Errorf("Expected reasoning in response for %s", providerName)
				}

				if _, hasConfidence := response["confidence"]; !hasConfidence {
					t.Errorf("Expected confidence score in response for %s", providerName)
				}
			})
		})
	}
}

// Test LLM fallback to basic rules
func TestLLMFallbackMechanism(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create config with invalid endpoint to trigger fallback
	config := moderation.LLMConfig{
		Provider:    "openai",
		APIKey:      "invalid-key",
		BaseURL:     "http://invalid-endpoint:9999",
		Model:       "gpt-4",
		MaxTokens:   500,
		Temperature: 0.1,
		Timeout:     1 * time.Second, // Short timeout
	}

	r := gin.New()
	r.Use(moderation.LLMModerationMiddleware(config))

	r.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Content approved by fallback"})
	})

	// Test that fallback works for safe content
	t.Run("fallback_safe_content", func(t *testing.T) {
		safeContent := map[string]interface{}{
			"message": "This is perfectly safe content",
			"author":  "user123",
		}

		jsonData, _ := json.Marshal(safeContent)
		req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200 for safe content with fallback, got %d", w.Code)
		}
	})

	// Test that fallback blocks unsafe content
	t.Run("fallback_unsafe_content", func(t *testing.T) {
		unsafeContent := map[string]interface{}{
			"message": "This is spam content",
			"author":  "spammer",
		}

		jsonData, _ := json.Marshal(unsafeContent)
		req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected 403 for unsafe content with fallback, got %d", w.Code)
		}
	})
}

// Test different content types with LLM
func TestLLMContentTypes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockServer := createMockLLMServer()
	defer mockServer.Close()

	config := moderation.LLMConfig{
		Provider:    "openai",
		APIKey:      "test-key",
		BaseURL:     mockServer.URL,
		Model:       "gpt-4",
		MaxTokens:   500,
		Temperature: 0.1,
		Timeout:     5 * time.Second,
	}

	r := gin.New()
	r.Use(moderation.LLMModerationMiddleware(config))

	r.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "LLM analysis passed"})
	})

	testCases := []struct {
		name        string
		content     interface{}
		expectSafe  bool
		description string
	}{
		{
			name: "blog_post",
			content: map[string]interface{}{
				"title":   "How to Build APIs in Go",
				"content": "Go is a great language for building scalable APIs...",
				"tags":    []string{"golang", "programming", "tutorial"},
			},
			expectSafe:  true,
			description: "Technical blog post should be safe",
		},
		{
			name: "product_review",
			content: map[string]interface{}{
				"rating": 5,
				"review": "Great product! Highly recommend to everyone.",
				"title":  "Excellent quality",
			},
			expectSafe:  true,
			description: "Positive product review should be safe",
		},
		{
			name: "social_post",
			content: map[string]interface{}{
				"content":    "Just finished reading an amazing book! 📚✨",
				"visibility": "public",
				"tags":       []string{"books", "reading"},
			},
			expectSafe:  true,
			description: "Social media post should be safe",
		},
		{
			name: "spam_promotion",
			content: map[string]interface{}{
				"message": "GET RICH QUICK! This is spam content with promotional links!",
				"urls":    []string{"http://spam-site.com"},
			},
			expectSafe:  false,
			description: "Spam promotional content should be blocked",
		},
		{
			name: "hate_speech",
			content: map[string]interface{}{
				"comment": "I hate this group of people and think they're terrible",
				"context": "forum discussion",
			},
			expectSafe:  false,
			description: "Hate speech should be blocked",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tc.content)
			req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			expectedStatus := http.StatusOK
			if !tc.expectSafe {
				expectedStatus = http.StatusForbidden
			}

			if w.Code != expectedStatus {
				t.Errorf("%s: Expected status %d, got %d. Response: %s",
					tc.description, expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}

// Test LLM response parsing
func TestLLMResponseParsing(t *testing.T) {
	testCases := []struct {
		name        string
		response    string
		expectSafe  bool
		expectError bool
		description string
	}{
		{
			name: "valid_json_safe",
			response: `{
				"is_safe": true,
				"confidence": 0.95,
				"categories": {"spam": 0.1},
				"reasoning": "Content appears safe"
			}`,
			expectSafe:  true,
			expectError: false,
			description: "Valid JSON with safe content",
		},
		{
			name: "valid_json_unsafe",
			response: `{
				"is_safe": false,
				"confidence": 0.85,
				"categories": {"spam": 0.9},
				"reasoning": "Content contains spam"
			}`,
			expectSafe:  false,
			expectError: false,
			description: "Valid JSON with unsafe content",
		},
		{
			name: "json_with_extra_text",
			response: `Here's my analysis:

			{
				"is_safe": true,
				"confidence": 0.92,
				"categories": {"spam": 0.05},
				"reasoning": "Content looks good"
			}

			That's my assessment.`,
			expectSafe:  true,
			expectError: false,
			description: "JSON embedded in extra text",
		},
		{
			name:        "invalid_json",
			response:    "This is not JSON at all",
			expectSafe:  false,
			expectError: true,
			description: "Invalid JSON should cause error",
		},
		{
			name:        "empty_response",
			response:    "",
			expectSafe:  false,
			expectError: true,
			description: "Empty response should cause error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := moderation.ParseLLMResponse(tc.response)

			if tc.expectError {
				if err == nil {
					t.Errorf("%s: Expected error but got none", tc.description)
				}
				return
			}

			if err != nil {
				t.Errorf("%s: Unexpected error: %v", tc.description, err)
				return
			}

			if result.IsSafe != tc.expectSafe {
				t.Errorf("%s: Expected is_safe=%t, got %t",
					tc.description, tc.expectSafe, result.IsSafe)
			}

			if result.Confidence <= 0 || result.Confidence > 1 {
				t.Errorf("%s: Invalid confidence score: %f",
					tc.description, result.Confidence)
			}
		})
	}
}

// Benchmark LLM moderation middleware
func BenchmarkLLMModerationMiddleware(b *testing.B) {
	gin.SetMode(gin.TestMode)

	mockServer := createMockLLMServer()
	defer mockServer.Close()

	config := moderation.LLMConfig{
		Provider:    "openai",
		APIKey:      "test-key",
		BaseURL:     mockServer.URL,
		Model:       "gpt-4",
		MaxTokens:   500,
		Temperature: 0.1,
		Timeout:     5 * time.Second,
	}

	r := gin.New()
	r.Use(moderation.LLMModerationMiddleware(config))

	r.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "OK"})
	})

	testContent := map[string]interface{}{
		"message": "This is a test message for benchmarking LLM moderation",
		"author":  "benchmark_user",
	}
	jsonData, _ := json.Marshal(testContent)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}
