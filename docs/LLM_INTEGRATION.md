# 🤖 LLM Integration Guide for Content Moderation

This guide explains how to integrate Large Language Models (LLMs) with our content moderation middleware for intelligent, AI-powered content analysis.

## 🎯 Overview

Our LLM integration supports multiple providers:
- **OpenAI GPT-4/GPT-3.5** - Industry-leading performance
- **Anthropic Claude** - Strong reasoning capabilities  
- **Hugging Face Models** - Open-source flexibility
- **Local Models** (Ollama) - Privacy and cost control

## 🚀 Quick Start

### 1. Basic Setup

```go
import "github.com/kart2405/API_Gateway/internal/middleware/moderation"

// Configure your LLM provider
config := moderation.LLMConfig{
    Provider:    "openai",
    APIKey:      "sk-your-api-key",
    BaseURL:     "https://api.openai.com/v1",
    Model:       "gpt-4",
    MaxTokens:   500,
    Temperature: 0.1,
    Timeout:     30 * time.Second,
}

// Apply to your Gin router
r.Use(moderation.LLMModerationMiddleware(config))
```

### 2. Provider-Specific Configuration

#### OpenAI Configuration
```go
openaiConfig := moderation.LLMConfig{
    Provider:    "openai",
    APIKey:      "sk-your-openai-key",
    BaseURL:     "https://api.openai.com/v1",
    Model:       "gpt-4",           // or "gpt-3.5-turbo"
    MaxTokens:   500,
    Temperature: 0.1,               // Low for consistent results
    Timeout:     30 * time.Second,
}
```

#### Anthropic Claude Configuration
```go
claudeConfig := moderation.LLMConfig{
    Provider:    "anthropic",
    APIKey:      "your-anthropic-key",
    BaseURL:     "https://api.anthropic.com/v1",
    Model:       "claude-3-sonnet-20240229",
    MaxTokens:   500,
    Temperature: 0.1,
    Timeout:     30 * time.Second,
}
```

#### Hugging Face Configuration
```go
hfConfig := moderation.LLMConfig{
    Provider:    "huggingface",
    APIKey:      "your-hf-token",
    BaseURL:     "https://api-inference.huggingface.co/models/microsoft/DialoGPT-large",
    Model:       "microsoft/DialoGPT-large",
    MaxTokens:   500,
    Temperature: 0.1,
    Timeout:     30 * time.Second,
}
```

#### Local Model (Ollama) Configuration
```go
localConfig := moderation.LLMConfig{
    Provider:    "local",
    APIKey:      "", // Not needed for local
    BaseURL:     "http://localhost:11434",
    Model:       "llama2:7b",       // or "mistral:7b", "codellama:7b"
    MaxTokens:   500,
    Temperature: 0.1,
    Timeout:     60 * time.Second,  // Longer timeout for local processing
}
```

## 🔧 Advanced Configuration

### Environment-Based Configuration
```go
func loadLLMConfig() moderation.LLMConfig {
    return moderation.LLMConfig{
        Provider:    os.Getenv("LLM_PROVIDER"),
        APIKey:      os.Getenv("LLM_API_KEY"),
        BaseURL:     os.Getenv("LLM_BASE_URL"),
        Model:       os.Getenv("LLM_MODEL"),
        MaxTokens:   getEnvInt("LLM_MAX_TOKENS", 500),
        Temperature: getEnvFloat("LLM_TEMPERATURE", 0.1),
        Timeout:     time.Duration(getEnvInt("LLM_TIMEOUT", 30)) * time.Second,
    }
}
```

### Multi-Provider Fallback
```go
func createFallbackMiddleware() gin.HandlerFunc {
    primaryConfig := moderation.LLMConfig{
        Provider: "openai",
        // ... config
    }
    
    fallbackConfig := moderation.LLMConfig{
        Provider: "local",
        // ... config  
    }
    
    return gin.HandlerFunc(func(c *gin.Context) {
        // Try primary provider
        if err := tryModeration(c, primaryConfig); err != nil {
            // Fall back to secondary provider
            tryModeration(c, fallbackConfig)
        }
    })
}
```

## 📊 Response Format

The LLM returns structured analysis:

```json
{
  "is_safe": true,
  "confidence": 0.95,
  "categories": {
    "spam": 0.1,
    "hate_speech": 0.05,
    "violence": 0.1,
    "harassment": 0.1,
    "adult_content": 0.05,
    "misinformation": 0.1
  },
  "reasoning": "Content appears to be a legitimate technical discussion",
  "suggestions": "Consider adding more context for clarity"
}
```

### Category Scores
- **0.0 - 0.3**: Very low risk
- **0.3 - 0.6**: Moderate risk  
- **0.6 - 0.8**: High risk
- **0.8 - 1.0**: Very high risk

## 🧪 Testing Your Integration

### 1. Run Basic Tests
```bash
# Test with mock LLM server
go test ./tests/llm_moderation_test.go -v

# Run benchmarks
go test ./tests/llm_moderation_test.go -bench=BenchmarkLLM
```

### 2. Test Different Providers
```go
func TestAllProviders(t *testing.T) {
    providers := []string{"openai", "anthropic", "huggingface", "local"}
    
    for _, provider := range providers {
        t.Run(provider, func(t *testing.T) {
            config := getConfigForProvider(provider)
            testModerationWithConfig(t, config)
        })
    }
}
```

### 3. Integration Testing
```bash
# Start local Ollama server (for local testing)
ollama serve

# Pull a model
ollama pull llama2:7b

# Run integration tests
go run examples/llm_moderation_example.go
```

## 🚀 Production Deployment

### 1. API Key Management
```bash
# Set environment variables
export LLM_PROVIDER="openai"
export LLM_API_KEY="sk-your-key"
export LLM_MODEL="gpt-4"
export LLM_MAX_TOKENS="500"
export LLM_TEMPERATURE="0.1"
export LLM_TIMEOUT="30"
```

### 2. Docker Configuration
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o api-gateway cmd/gateway/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/api-gateway .

# LLM configuration
ENV LLM_PROVIDER=openai
ENV LLM_MODEL=gpt-4
ENV LLM_MAX_TOKENS=500
ENV LLM_TEMPERATURE=0.1
ENV LLM_TIMEOUT=30

CMD ["./api-gateway"]
```

### 3. Kubernetes Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-gateway-llm
spec:
  replicas: 3
  selector:
    matchLabels:
      app: api-gateway-llm
  template:
    metadata:
      labels:
        app: api-gateway-llm
    spec:
      containers:
      - name: api-gateway
        image: your-registry/api-gateway:latest
        env:
        - name: LLM_PROVIDER
          value: "openai"
        - name: LLM_API_KEY
          valueFrom:
            secretKeyRef:
              name: llm-secrets
              key: api-key
        - name: LLM_MODEL
          value: "gpt-4"
        ports:
        - containerPort: 8080
```

## 📈 Performance Optimization

### 1. Caching Responses
```go
type CachedModerationResult struct {
    Result    *moderation.ModerationResult
    ExpiresAt time.Time
}

var cache = make(map[string]*CachedModerationResult)

func getCachedResult(contentHash string) (*moderation.ModerationResult, bool) {
    if cached, exists := cache[contentHash]; exists {
        if time.Now().Before(cached.ExpiresAt) {
            return cached.Result, true
        }
        delete(cache, contentHash)
    }
    return nil, false
}
```

### 2. Async Processing
```go
func AsyncModerationMiddleware(config moderation.LLMConfig) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Allow request to proceed
        c.Next()
        
        // Process moderation asynchronously
        go func() {
            body, _ := c.Get("requestBody")
            result, _ := moderation.AnalyzeLLMContent(body.(string), config)
            
            // Store result for future use or alerting
            storeModerationResult(result)
        }()
    }
}
```

### 3. Rate Limiting
```go
func RateLimitedLLMMiddleware(config moderation.LLMConfig) gin.HandlerFunc {
    limiter := rate.NewLimiter(10, 20) // 10 req/sec, burst of 20
    
    return func(c *gin.Context) {
        if !limiter.Allow() {
            // Fall back to basic moderation
            basicModerationMiddleware(c)
            return
        }
        
        // Use LLM moderation
        llmModerationMiddleware(config)(c)
    }
}
```

## 🔍 Monitoring and Logging

### 1. Metrics Collection
```go
var (
    llmRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "llm_moderation_requests_total",
            Help: "Total number of LLM moderation requests",
        },
        []string{"provider", "model", "result"},
    )
    
    llmRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "llm_moderation_request_duration_seconds",
            Help: "LLM moderation request duration",
        },
        []string{"provider", "model"},
    )
)
```

### 2. Structured Logging
```go
func logModerationResult(content string, result *moderation.ModerationResult, provider string) {
    log.WithFields(logrus.Fields{
        "provider":     provider,
        "is_safe":      result.IsSafe,
        "confidence":   result.Confidence,
        "content_hash": hashContent(content),
        "categories":   result.Categories,
    }).Info("LLM moderation completed")
}
```

## 🛠️ Troubleshooting

### Common Issues

1. **API Key Issues**
   ```bash
   # Test API key
   curl -H "Authorization: Bearer $LLM_API_KEY" \
        https://api.openai.com/v1/models
   ```

2. **Timeout Errors**
   ```go
   // Increase timeout for complex content
   config.Timeout = 60 * time.Second
   ```

3. **Rate Limiting**
   ```go
   // Implement exponential backoff
   func retryWithBackoff(fn func() error, maxRetries int) error {
       for i := 0; i < maxRetries; i++ {
           if err := fn(); err == nil {
               return nil
           }
           time.Sleep(time.Duration(math.Pow(2, float64(i))) * time.Second)
       }
       return fmt.Errorf("max retries exceeded")
   }
   ```

4. **Local Model Issues**
   ```bash
   # Check Ollama status
   ollama list
   
   # Pull model if missing
   ollama pull llama2:7b
   
   # Check logs
   ollama logs
   ```

## 📚 Best Practices

### 1. Content Preparation
- Clean and normalize text before sending to LLM
- Remove sensitive information (PII, credentials)
- Limit content length to avoid token limits

### 2. Error Handling
- Always implement fallback mechanisms
- Log all errors for debugging
- Provide meaningful error messages to users

### 3. Security
- Never log API keys or sensitive content
- Use environment variables for configuration
- Implement proper access controls

### 4. Cost Management
- Monitor token usage and costs
- Implement caching for repeated content
- Use appropriate models for your use case

## 🔮 Future Enhancements

- **Custom Model Training**: Fine-tune models for your specific use case
- **Multi-Modal Support**: Analyze images, videos, and audio content
- **Real-Time Streaming**: Process content as it's being typed
- **Advanced Analytics**: Detailed reporting and trend analysis
- **A/B Testing**: Compare different models and configurations

## 📞 Support

For issues or questions:
1. Check the troubleshooting section above
2. Review test examples in `tests/llm_moderation_test.go`
3. Run the example server: `go run examples/llm_moderation_example.go`
4. Open an issue with detailed logs and configuration
