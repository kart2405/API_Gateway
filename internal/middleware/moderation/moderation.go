package moderation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ModerationResponse represents the response from the AI moderation service
type ModerationResponse struct {
	IsSafe bool   `json:"is_safe"`
	Reason string `json:"reason,omitempty"`
}

// ModerationRequest represents the request payload sent to the AI moderation service
type ModerationRequest struct {
	Content string `json:"content"`
}

// ModerationMiddleware creates a middleware for AI-powered content moderation
// This middleware intercepts requests, analyzes the content for policy violations,
// and blocks unsafe content before it reaches downstream handlers
func ModerationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only process requests with body content (POST, PUT, PATCH)
		if c.Request.Method != http.MethodPost &&
			c.Request.Method != http.MethodPut &&
			c.Request.Method != http.MethodPatch {
			c.Next()
			return
		}

		// Safely read the request body
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			c.Abort()
			return
		}

		// Restore the request body so downstream handlers can read it
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Skip moderation if body is empty
		if len(bodyBytes) == 0 {
			c.Next()
			return
		}

		// Convert body to string for content analysis
		content := string(bodyBytes)

		// Call the AI moderation service to analyze the content
		isSafe, err := callModerationService(content)
		if err != nil {
			// Log the error but don't block the request if moderation service is down
			// In production, you might want to implement different error handling strategies
			fmt.Printf("Moderation service error: %v\n", err)

			// For now, we'll be strict and block on errors
			// You could change this to allow requests through on service failures
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Content moderation service temporarily unavailable",
			})
			c.Abort()
			return
		}

		// Block unsafe content
		if !isSafe {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Content violates our community guidelines",
			})
			c.Abort()
			return
		}

		// Content is safe, proceed to next handler
		c.Next()
	}
}

// callModerationService simulates making an HTTP POST request to an external AI moderation service
// This is a placeholder implementation that will be replaced with actual service integration
//
// Parameters:
//   - content: The request body content to be analyzed
//
// Returns:
//   - bool: true if content is safe, false if it violates guidelines
//   - error: any error that occurred during the moderation check
func callModerationService(content string) (bool, error) {
	// Placeholder implementation - replace with actual AI moderation service call
	// For now, we'll implement some basic rules as a demonstration

	// In a real implementation, you would create a moderation request payload:
	// moderationReq := ModerationRequest{
	//     Content: content,
	// }
	//
	// Then you would:
	// 1. Marshal the request to JSON
	// 2. Make an HTTP POST to the moderation service endpoint
	// 3. Parse the response
	// 4. Return the safety assessment

	// Example of what the real implementation might look like:
	/*
		reqBody, err := json.Marshal(moderationReq)
		if err != nil {
			return false, fmt.Errorf("failed to marshal moderation request: %w", err)
		}

		client := &http.Client{
			Timeout: 5 * time.Second, // Set reasonable timeout
		}

		resp, err := client.Post(
			"http://ai-moderation-service:5000/moderate",
			"application/json",
			bytes.NewBuffer(reqBody),
		)
		if err != nil {
			return false, fmt.Errorf("failed to call moderation service: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return false, fmt.Errorf("moderation service returned status: %d", resp.StatusCode)
		}

		var moderationResp ModerationResponse
		if err := json.NewDecoder(resp.Body).Decode(&moderationResp); err != nil {
			return false, fmt.Errorf("failed to decode moderation response: %w", err)
		}

		return moderationResp.IsSafe, nil
	*/

	// For demonstration purposes, implement some basic content filtering
	// This is a simplified placeholder - real AI moderation would be much more sophisticated

	// Simulate network delay
	time.Sleep(10 * time.Millisecond)

	// Parse JSON to check for specific patterns (very basic example)
	var jsonContent map[string]interface{}
	if err := json.Unmarshal([]byte(content), &jsonContent); err != nil {
		// If it's not JSON, treat as plain text
		return isContentSafe(content), nil
	}

	// Check all string values in the JSON for unsafe content
	return isJSONContentSafe(jsonContent), nil
}

// isContentSafe performs basic content safety checks on plain text
// This is a placeholder implementation for demonstration
func isContentSafe(content string) bool {
	// Comprehensive list of prohibited terms for spam and harmful content detection
	prohibitedTerms := []string{
		// Basic spam terms
		"spam", "spammer", "spamming",

		// Promotional/commercial spam
		"buy now", "click here", "limited time", "act now", "urgent", "hurry",
		"free money", "make money fast", "get rich quick", "work from home",
		"guaranteed income", "no experience needed", "easy money",
		"discount", "sale", "offer expires", "special deal", "lowest price",
		"casino", "gambling", "lottery", "winner", "congratulations you won",

		// Phishing and scam terms
		"verify account", "update payment", "suspended account", "click to verify",
		"confirm identity", "security alert", "account locked", "immediate action",
		"wire transfer", "western union", "money gram", "bitcoin payment",
		"inheritance", "lottery winner", "beneficiary", "transfer funds",
		"nigerian prince", "unclaimed money", "tax refund",

		// Adult/inappropriate content
		"adult content", "xxx", "porn", "sexual", "dating site", "hookup",
		"viagra", "cialis", "enhancement", "enlargement",

		// Repetitive/low quality content
		"lorem ipsum", "test test test", "aaaaaa", "click click click",
		"like and subscribe", "follow for follow", "f4f", "l4l",

		// Hate speech and harassment
		"hate", "violence", "abuse", "kill", "die", "stupid", "idiot",
		"harassment", "bully", "threat", "revenge", "destroy",

		// Misinformation indicators
		"fake news", "conspiracy", "hoax", "scam", "fraud", "cheat",
		"illegal", "stolen", "hacked", "cracked", "pirated",

		// Social media spam
		"follow me", "check my profile", "link in bio", "dm me",
		"add me", "friend request", "social media", "instagram",
		"facebook", "twitter", "tiktok", "youtube channel",

		// Cryptocurrency spam
		"crypto", "bitcoin", "ethereum", "nft", "trading bot", "pump",
		"investment opportunity", "trading signals", "forex",

		// MLM and pyramid schemes
		"mlm", "network marketing", "pyramid", "downline", "upline",
		"residual income", "passive income", "financial freedom",

		// Medical spam
		"weight loss", "diet pills", "miracle cure", "health supplement",
		"medical breakthrough", "doctor recommended",

		// Tech support scams
		"tech support", "virus detected", "computer infected", "call now",
		"microsoft support", "apple support", "google support",

		// Generic spam patterns
		"act fast", "limited offer", "exclusive deal", "one time only",
		"don't miss out", "last chance", "expires soon", "while supplies last",
	}

	contentLower := strings.ToLower(content)
	for _, term := range prohibitedTerms {
		if strings.Contains(contentLower, term) {
			return false
		}
	}

	return true
}

// isJSONContentSafe recursively checks JSON content for safety
// This is a placeholder implementation for demonstration
func isJSONContentSafe(data interface{}) bool {
	switch v := data.(type) {
	case string:
		return isContentSafe(v)
	case map[string]interface{}:
		for _, value := range v {
			if !isJSONContentSafe(value) {
				return false
			}
		}
	case []interface{}:
		for _, item := range v {
			if !isJSONContentSafe(item) {
				return false
			}
		}
	}
	return true
}
