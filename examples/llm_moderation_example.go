package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kart2405/API_Gateway/internal/middleware/moderation"
)

func main() {
	// Configuration for different LLM providers
	configs := map[string]moderation.LLMConfig{
		"openai": {
			Provider:    "openai",
			APIKey:      "sk-your-openai-api-key",
			BaseURL:     "https://api.openai.com/v1",
			Model:       "gpt-4",
			MaxTokens:   500,
			Temperature: 0.1,
			Timeout:     30 * time.Second,
		},
		"claude": {
			Provider:    "anthropic",
			APIKey:      "your-anthropic-api-key",
			BaseURL:     "https://api.anthropic.com/v1",
			Model:       "claude-3-sonnet-20240229",
			MaxTokens:   500,
			Temperature: 0.1,
			Timeout:     30 * time.Second,
		},
		"huggingface": {
			Provider:    "huggingface",
			APIKey:      "your-hf-token",
			BaseURL:     "https://api-inference.huggingface.co/models/microsoft/DialoGPT-large",
			Model:       "microsoft/DialoGPT-large",
			MaxTokens:   500,
			Temperature: 0.1,
			Timeout:     30 * time.Second,
		},
		"local": {
			Provider:    "local",
			APIKey:      "",                       // Not needed for local
			BaseURL:     "http://localhost:11434", // Ollama default
			Model:       "llama2:7b",
			MaxTokens:   500,
			Temperature: 0.1,
			Timeout:     60 * time.Second,
		},
	}

	// Choose your LLM provider
	selectedConfig := configs["openai"] // Change this to test different providers

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Apply LLM-powered moderation middleware
	r.Use(moderation.LLMModerationMiddleware(selectedConfig))

	// Example endpoints
	setupExampleEndpoints(r)

	// Start server
	log.Println("🤖 Starting LLM-powered moderation server on :8080")
	log.Printf("📡 Using provider: %s with model: %s", selectedConfig.Provider, selectedConfig.Model)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupExampleEndpoints(r *gin.Engine) {
	// Blog posts endpoint
	r.POST("/posts", func(c *gin.Context) {
		var post struct {
			Title   string `json:"title" binding:"required"`
			Content string `json:"content" binding:"required"`
			Author  string `json:"author" binding:"required"`
		}

		if err := c.ShouldBindJSON(&post); err != nil {
			c.JSON(400, gin.H{"error": "Invalid post data"})
			return
		}

		c.JSON(201, gin.H{
			"message": "Post created successfully! LLM verified it's safe.",
			"post": gin.H{
				"id":     12345,
				"title":  post.Title,
				"author": post.Author,
				"status": "published",
			},
		})
	})

	// Comments endpoint
	r.POST("/comments", func(c *gin.Context) {
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
			"message": "Comment posted successfully! LLM moderation passed.",
			"comment": gin.H{
				"id":      67890,
				"post_id": comment.PostID,
				"author":  comment.Author,
				"status":  "approved",
			},
		})
	})

	// User reviews endpoint
	r.POST("/reviews", func(c *gin.Context) {
		var review struct {
			ProductID int    `json:"product_id" binding:"required"`
			Rating    int    `json:"rating" binding:"required,min=1,max=5"`
			Review    string `json:"review" binding:"required"`
			Title     string `json:"title"`
		}

		if err := c.ShouldBindJSON(&review); err != nil {
			c.JSON(400, gin.H{"error": "Invalid review data"})
			return
		}

		c.JSON(201, gin.H{
			"message": "Review submitted successfully! AI moderation approved.",
			"review": gin.H{
				"id":         54321,
				"product_id": review.ProductID,
				"rating":     review.Rating,
				"status":     "published",
			},
		})
	})

	// Social media posts
	r.POST("/social/posts", func(c *gin.Context) {
		var socialPost struct {
			Content    string   `json:"content" binding:"required"`
			Tags       []string `json:"tags"`
			Visibility string   `json:"visibility"`
		}

		if err := c.ShouldBindJSON(&socialPost); err != nil {
			c.JSON(400, gin.H{"error": "Invalid social post data"})
			return
		}

		c.JSON(201, gin.H{
			"message": "Social post published! LLM content analysis passed.",
			"post": gin.H{
				"id":         98765,
				"content":    socialPost.Content,
				"visibility": socialPost.Visibility,
				"status":     "live",
			},
		})
	})

	// Test endpoint for different content types
	r.POST("/test/content", func(c *gin.Context) {
		var testContent struct {
			Type    string `json:"type"` // "safe", "spam", "hate", "violence"
			Content string `json:"content"`
		}

		if err := c.ShouldBindJSON(&testContent); err != nil {
			c.JSON(400, gin.H{"error": "Invalid test data"})
			return
		}

		c.JSON(200, gin.H{
			"message": "Content passed LLM moderation!",
			"analysis": gin.H{
				"type":         testContent.Type,
				"content_safe": true,
				"ai_verified":  true,
			},
		})
	})

	// Health check (bypasses moderation - GET request)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "llm-moderation-api",
		})
	})

	// API documentation
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"name":        "LLM-Powered Content Moderation API",
			"version":     "2.0.0",
			"description": "API Gateway with AI-powered content moderation using various LLM providers",
			"endpoints": gin.H{
				"POST /posts":        "Create blog post (with LLM moderation)",
				"POST /comments":     "Post comment (with LLM moderation)",
				"POST /reviews":      "Submit product review (with LLM moderation)",
				"POST /social/posts": "Create social media post (with LLM moderation)",
				"POST /test/content": "Test content moderation",
				"GET  /health":       "Health check",
			},
			"features": []string{
				"Multi-provider LLM support (OpenAI, Anthropic, Hugging Face, Local)",
				"Intelligent content analysis with confidence scores",
				"Detailed reasoning for moderation decisions",
				"Fallback to rule-based moderation if LLM fails",
				"Support for JSON and plain text content",
				"Real-time content safety assessment",
			},
		})
	})
}
