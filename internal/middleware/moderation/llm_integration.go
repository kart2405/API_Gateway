package moderation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// LLMConfig holds configuration for different LLM providers
type LLMConfig struct {
	Provider    string        `json:"provider"` // "openai", "anthropic", "huggingface", "local"
	APIKey      string        `json:"api_key"`
	BaseURL     string        `json:"base_url"`
	Model       string        `json:"model"`
	MaxTokens   int           `json:"max_tokens"`
	Temperature float64       `json:"temperature"`
	Timeout     time.Duration `json:"timeout"`
}

// ModerationResult represents the LLM's moderation decision
type ModerationResult struct {
	IsSafe      bool               `json:"is_safe"`
	Confidence  float64            `json:"confidence"`
	Categories  map[string]float64 `json:"categories"`
	Reasoning   string             `json:"reasoning"`
	Suggestions string             `json:"suggestions,omitempty"`
}

// OpenAI Integration
type OpenAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Enhanced moderation middleware with LLM integration
func LLMModerationMiddleware(config LLMConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only process content-creating requests
		if c.Request.Method != http.MethodPost &&
			c.Request.Method != http.MethodPut &&
			c.Request.Method != http.MethodPatch {
			c.Next()
			return
		}

		// Read and preserve request body
		bodyBytes, err := readAndPreserveBody(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			c.Abort()
			return
		}

		// Skip empty bodies
		if len(bodyBytes) == 0 {
			c.Next()
			return
		}

		// Analyze content with LLM
		result, err := analyzeLLMContent(string(bodyBytes), config)
		if err != nil {
			// Fallback to basic rules if LLM fails
			fmt.Printf("LLM moderation failed, using fallback: %v\n", err)
			isSafe := basicContentCheck(string(bodyBytes))
			if !isSafe {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "Content violates our community guidelines",
				})
				c.Abort()
				return
			}
		} else {
			// Use LLM result
			if !result.IsSafe {
				c.JSON(http.StatusForbidden, gin.H{
					"error":       "Content violates our community guidelines",
					"reasoning":   result.Reasoning,
					"confidence":  result.Confidence,
					"suggestions": result.Suggestions,
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// OpenAI GPT-4 Integration
func moderateWithOpenAI(content string, config LLMConfig) (*ModerationResult, error) {
	prompt := buildModerationPrompt(content)

	request := OpenAIRequest{
		Model: config.Model,
		Messages: []Message{
			{
				Role:    "system",
				Content: "You are a content moderation AI. Analyze content for safety and provide detailed feedback in JSON format.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   config.MaxTokens,
		Temperature: config.Temperature,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	client := &http.Client{Timeout: config.Timeout}
	req, err := http.NewRequest("POST", config.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API returned status: %d", resp.StatusCode)
	}

	var openAIResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from OpenAI")
	}

	// Parse the LLM response
	return ParseLLMResponse(openAIResp.Choices[0].Message.Content)
}

// Anthropic Claude Integration
func moderateWithClaude(content string, config LLMConfig) (*ModerationResult, error) {
	type ClaudeRequest struct {
		Model     string    `json:"model"`
		MaxTokens int       `json:"max_tokens"`
		Messages  []Message `json:"messages"`
	}

	prompt := buildModerationPrompt(content)

	request := ClaudeRequest{
		Model:     config.Model,
		MaxTokens: config.MaxTokens,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	client := &http.Client{Timeout: config.Timeout}
	req, err := http.NewRequest("POST", config.BaseURL+"/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Claude API: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract content from Claude response
	content_text := ""
	if content_arr, ok := result["content"].([]interface{}); ok && len(content_arr) > 0 {
		if text_obj, ok := content_arr[0].(map[string]interface{}); ok {
			content_text = text_obj["text"].(string)
		}
	}

	return ParseLLMResponse(content_text)
}

// Hugging Face Integration (for open-source models)
func moderateWithHuggingFace(content string, config LLMConfig) (*ModerationResult, error) {
	type HFRequest struct {
		Inputs     string                 `json:"inputs"`
		Parameters map[string]interface{} `json:"parameters"`
	}

	prompt := buildModerationPrompt(content)

	request := HFRequest{
		Inputs: prompt,
		Parameters: map[string]interface{}{
			"max_new_tokens":   config.MaxTokens,
			"temperature":      config.Temperature,
			"return_full_text": false,
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	client := &http.Client{Timeout: config.Timeout}
	req, err := http.NewRequest("POST", config.BaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Hugging Face API: %w", err)
	}
	defer resp.Body.Close()

	var result []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no results returned")
	}

	generatedText := result[0]["generated_text"].(string)
	return ParseLLMResponse(generatedText)
}

// Local model integration (using Ollama or similar)
func moderateWithLocalModel(content string, config LLMConfig) (*ModerationResult, error) {
	type LocalRequest struct {
		Model  string `json:"model"`
		Prompt string `json:"prompt"`
		Stream bool   `json:"stream"`
	}

	prompt := buildModerationPrompt(content)

	request := LocalRequest{
		Model:  config.Model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	client := &http.Client{Timeout: config.Timeout}
	req, err := http.NewRequest("POST", config.BaseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call local model: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	response := result["response"].(string)
	return ParseLLMResponse(response)
}

// Build comprehensive moderation prompt
func buildModerationPrompt(content string) string {
	return fmt.Sprintf(`
Analyze the following content for safety and policy violations. Consider these categories:
- Spam: Repetitive, promotional, or irrelevant content
- Hate Speech: Content targeting individuals or groups based on identity
- Violence: Threats, graphic violence, or incitement to harm
- Harassment: Bullying, stalking, or targeted abuse
- Adult Content: Sexual or explicit material inappropriate for general audiences
- Misinformation: False or misleading information that could cause harm

Content to analyze:
"%s"

Respond in this exact JSON format:
{
  "is_safe": true/false,
  "confidence": 0.0-1.0,
  "categories": {
    "spam": 0.0-1.0,
    "hate_speech": 0.0-1.0,
    "violence": 0.0-1.0,
    "harassment": 0.0-1.0,
    "adult_content": 0.0-1.0,
    "misinformation": 0.0-1.0
  },
  "reasoning": "Brief explanation of the decision",
  "suggestions": "How to improve the content if unsafe (optional)"
}`, content)
}

// ParseLLMResponse parses LLM response into structured result
func ParseLLMResponse(response string) (*ModerationResult, error) {
	// Extract JSON from response (LLMs sometimes add extra text)
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}") + 1

	if start == -1 || end == 0 {
		return nil, fmt.Errorf("no JSON found in response")
	}

	jsonStr := response[start:end]

	var result ModerationResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return &result, nil
}

// Main LLM analysis function that routes to appropriate provider
func analyzeLLMContent(content string, config LLMConfig) (*ModerationResult, error) {
	switch strings.ToLower(config.Provider) {
	case "openai":
		return moderateWithOpenAI(content, config)
	case "anthropic", "claude":
		return moderateWithClaude(content, config)
	case "huggingface":
		return moderateWithHuggingFace(content, config)
	case "local", "ollama":
		return moderateWithLocalModel(content, config)
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", config.Provider)
	}
}

// Helper function to read and preserve request body
func readAndPreserveBody(c *gin.Context) ([]byte, error) {
	bodyBytes, err := c.GetRawData()
	if err != nil {
		return nil, err
	}

	// Restore body for downstream handlers
	c.Request.Body = http.NoBody
	if len(bodyBytes) > 0 {
		c.Request.Body = &readCloser{bytes.NewReader(bodyBytes)}
	}

	return bodyBytes, nil
}

type readCloser struct {
	*bytes.Reader
}

func (rc *readCloser) Close() error {
	return nil
}

// Fallback basic content check
func basicContentCheck(content string) bool {
	// Comprehensive list of prohibited terms for fallback spam detection
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
