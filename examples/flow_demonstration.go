package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kart2405/API_Gateway/internal/middleware/moderation"
)

// Demonstration of how the LLM moderation flow works
func main() {
	fmt.Println("🤖 LLM Content Moderation Flow Demonstration")
	fmt.Println("============================================")

	// Step 1: Setup mock LLM server to see the flow
	mockLLMServer := createDemoLLMServer()
	defer mockLLMServer.Close()

	// Step 2: Configure LLM middleware
	config := moderation.LLMConfig{
		Provider:    "openai",
		APIKey:      "demo-key",
		BaseURL:     mockLLMServer.URL,
		Model:       "gpt-4",
		MaxTokens:   500,
		Temperature: 0.1,
		Timeout:     5 * time.Second,
	}

	// Step 3: Create API with LLM moderation
	r := gin.New()
	r.Use(gin.LoggerWithFormatter(customLogFormat))
	r.Use(moderation.LLMModerationMiddleware(config))

	// Step 4: Add endpoints that will be protected
	setupDemoEndpoints(r)

	// Step 5: Start server and demonstrate the flow
	go func() {
		log.Println("🚀 Demo server starting on :8080")
		r.Run(":8080")
	}()

	// Wait for server to start
	time.Sleep(1 * time.Second)

	// Step 6: Demonstrate different scenarios
	demonstrateFlow()
}

// Custom log format to show the flow clearly
func customLogFormat(param gin.LogFormatterParams) string {
	return fmt.Sprintf("🌊 [%s] %s %s → %d (%v)\n",
		param.TimeStamp.Format("15:04:05"),
		param.Method,
		param.Path,
		param.StatusCode,
		param.Latency,
	)
}

// Mock LLM server that shows what's happening behind the scenes
func createDemoLLMServer() *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("\n🤖 LLM SERVER: Received moderation request")

		// Parse the request
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)

		// Extract the content from messages
		messages := req["messages"].([]interface{})
		userMessage := messages[1].(map[string]interface{})
		prompt := userMessage["content"].(string)

		fmt.Printf("📝 CONTENT TO ANALYZE: %s\n", extractContentFromPrompt(prompt))

		// Simulate LLM analysis
		fmt.Println("🔍 LLM THINKING: Analyzing content for safety violations...")
		time.Sleep(100 * time.Millisecond) // Simulate processing time

		// Determine if content is safe
		isSafe, reasoning, categories := analyzeMockContent(prompt)

		fmt.Printf("🎯 LLM DECISION: is_safe=%t\n", isSafe)
		fmt.Printf("💭 LLM REASONING: %s\n", reasoning)

		// Create response
		response := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"content": fmt.Sprintf(`{
							"is_safe": %t,
							"confidence": 0.90,
							"categories": %s,
							"reasoning": "%s",
							"suggestions": "Consider rephrasing for better clarity"
						}`, isSafe, categories, reasoning),
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

		fmt.Println("📤 LLM SERVER: Response sent back to middleware")
	})

	server := &http.Server{
		Addr:    ":9999",
		Handler: mux,
	}

	go server.ListenAndServe()
	return server
}

// Extract content from the LLM prompt for demonstration
func extractContentFromPrompt(prompt string) string {
	// Simple extraction for demo purposes
	start := `Content: "`
	end := `"`

	startIdx := len(start)
	if idx := bytes.Index([]byte(prompt), []byte(start)); idx != -1 {
		startIdx = idx + len(start)
	}

	endIdx := len(prompt)
	if idx := bytes.Index([]byte(prompt[startIdx:]), []byte(end)); idx != -1 {
		endIdx = startIdx + idx
	}

	if startIdx < endIdx {
		return prompt[startIdx:endIdx]
	}
	return "Content extraction failed"
}

// Mock content analysis logic
func analyzeMockContent(prompt string) (bool, string, string) {
	content := extractContentFromPrompt(prompt)
	contentLower := strings.ToLower(content)

	// Check for unsafe patterns
	if strings.Contains(contentLower, "spam") {
		return false, "Content contains promotional spam patterns",
			`{"spam": 0.9, "hate_speech": 0.1, "violence": 0.1}`
	}

	if strings.Contains(contentLower, "hate") {
		return false, "Content contains hate speech elements",
			`{"spam": 0.1, "hate_speech": 0.9, "violence": 0.1}`
	}

	if strings.Contains(contentLower, "violence") {
		return false, "Content contains violent language",
			`{"spam": 0.1, "hate_speech": 0.1, "violence": 0.9}`
	}

	// Content is safe
	return true, "Content appears safe and appropriate",
		`{"spam": 0.1, "hate_speech": 0.05, "violence": 0.05}`
}

// Setup demo endpoints
func setupDemoEndpoints(r *gin.Engine) {
	r.POST("/posts", func(c *gin.Context) {
		fmt.Println("✅ HANDLER: Post creation handler reached - content was approved!")

		var post struct {
			Title   string `json:"title"`
			Content string `json:"content"`
		}

		c.ShouldBindJSON(&post)

		c.JSON(200, gin.H{
			"message": "Post created successfully!",
			"post": gin.H{
				"id":     12345,
				"title":  post.Title,
				"status": "published",
			},
		})
	})

	r.POST("/comments", func(c *gin.Context) {
		fmt.Println("✅ HANDLER: Comment handler reached - content was approved!")

		c.JSON(200, gin.H{
			"message":    "Comment posted successfully!",
			"comment_id": 67890,
		})
	})

	// GET endpoint (bypasses moderation)
	r.GET("/posts", func(c *gin.Context) {
		fmt.Println("⚡ HANDLER: GET request - bypassed moderation entirely!")

		c.JSON(200, gin.H{
			"posts": []map[string]interface{}{
				{"id": 1, "title": "Sample Post", "content": "Sample content"},
			},
		})
	})
}

// Demonstrate the complete flow with different scenarios
func demonstrateFlow() {
	fmt.Println("\n🎬 DEMONSTRATION: Testing different content scenarios")
	fmt.Println("==================================================")

	scenarios := []struct {
		name        string
		method      string
		endpoint    string
		content     map[string]interface{}
		description string
	}{
		{
			name:     "Safe Content",
			method:   "POST",
			endpoint: "/posts",
			content: map[string]interface{}{
				"title":   "How to Build APIs in Go",
				"content": "Go is a great programming language for building scalable APIs...",
			},
			description: "✅ This should pass moderation and create the post",
		},
		{
			name:     "Spam Content",
			method:   "POST",
			endpoint: "/posts",
			content: map[string]interface{}{
				"title":   "GET RICH QUICK!",
				"content": "This is spam content with promotional offers!",
			},
			description: "🚫 This should be blocked by LLM moderation",
		},
		{
			name:     "Hate Speech",
			method:   "POST",
			endpoint: "/comments",
			content: map[string]interface{}{
				"post_id": 123,
				"content": "I hate this article and the author is terrible",
			},
			description: "🚫 This should be blocked for hate speech",
		},
		{
			name:        "GET Request",
			method:      "GET",
			endpoint:    "/posts",
			content:     nil,
			description: "⚡ This should bypass moderation entirely",
		},
	}

	for i, scenario := range scenarios {
		fmt.Printf("\n📋 SCENARIO %d: %s\n", i+1, scenario.name)
		fmt.Printf("📝 DESCRIPTION: %s\n", scenario.description)
		fmt.Printf("🎯 REQUEST: %s %s\n", scenario.method, scenario.endpoint)

		if scenario.content != nil {
			contentBytes, _ := json.MarshalIndent(scenario.content, "", "  ")
			fmt.Printf("📄 CONTENT: %s\n", string(contentBytes))
		}

		// Make the request
		makeTestRequest(scenario.method, scenario.endpoint, scenario.content)

		fmt.Println("─────────────────────────────────────────────────")
		time.Sleep(1 * time.Second) // Pause between scenarios
	}

	fmt.Println("\n🎉 DEMONSTRATION COMPLETE!")
	fmt.Println("You can see the complete flow from request → LLM → decision → response")
}

// Make test requests to demonstrate the flow
func makeTestRequest(method, endpoint string, content map[string]interface{}) {
	var body []byte
	if content != nil {
		body, _ = json.Marshal(content)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	req, _ := http.NewRequest(method, "http://localhost:8080"+endpoint, bytes.NewBuffer(body))

	if content != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	fmt.Println("📡 SENDING REQUEST...")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ REQUEST FAILED: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	fmt.Printf("📨 RESPONSE STATUS: %d\n", resp.StatusCode)

	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		fmt.Printf("✅ SUCCESS: %v\n", response["message"])
	} else {
		fmt.Printf("🚫 BLOCKED: %v\n", response["error"])
		if reasoning, ok := response["reasoning"]; ok {
			fmt.Printf("💭 AI REASONING: %v\n", reasoning)
		}
	}
}
