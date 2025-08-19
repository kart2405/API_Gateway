package cache

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kart2405/API_Gateway/internal/config"
	"github.com/kart2405/API_Gateway/internal/middleware/logging"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// CacheEntry represents a cache entry
type CacheEntry struct {
	Data      interface{}   `json:"data"`
	Timestamp time.Time     `json:"timestamp"`
	TTL       time.Duration `json:"ttl"`
}

// CacheService manages caching operations
type CacheService struct {
	redisClient *redis.Client
	memoryCache map[string]CacheEntry
	mutex       sync.RWMutex
	enabled     bool
}

// NewCacheService creates a new cache service
func NewCacheService() *CacheService {
	return &CacheService{
		redisClient: config.GetRedisClient(),
		memoryCache: make(map[string]CacheEntry),
		enabled:     true,
	}
}

// generateKey generates a cache key from request
func (cs *CacheService) generateKey(c *gin.Context) string {
	// Create a unique key based on method, path, and query parameters
	key := fmt.Sprintf("%s:%s", c.Request.Method, c.Request.URL.Path)

	// Add query parameters to key
	if c.Request.URL.RawQuery != "" {
		key += "?" + c.Request.URL.RawQuery
	}

	// Add user ID if available
	if userID, exists := c.Get("userID"); exists {
		key += fmt.Sprintf(":user:%v", userID)
	}

	// Create MD5 hash of the key for consistent length
	hash := md5.Sum([]byte(key))
	return fmt.Sprintf("cache:%x", hash)
}

// Set sets a value in cache
func (cs *CacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if !cs.enabled {
		return nil
	}

	entry := CacheEntry{
		Data:      value,
		Timestamp: time.Now(),
		TTL:       ttl,
	}

	// Try Redis first
	if cs.redisClient != nil {
		data, err := json.Marshal(entry)
		if err != nil {
			return fmt.Errorf("failed to marshal cache entry: %w", err)
		}

		err = cs.redisClient.Set(ctx, key, data, ttl).Err()
		if err == nil {
			return nil
		}

		// Log Redis error but continue with memory cache
		logger := logging.GetLogger()
		logger.Warn("Redis cache set failed, falling back to memory cache",
			zap.String("key", key),
			zap.Error(err),
		)
	}

	// Fallback to memory cache
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	cs.memoryCache[key] = entry

	return nil
}

// Get gets a value from cache
func (cs *CacheService) Get(ctx context.Context, key string) (interface{}, bool) {
	if !cs.enabled {
		return nil, false
	}

	// Try Redis first
	if cs.redisClient != nil {
		data, err := cs.redisClient.Get(ctx, key).Result()
		if err == nil {
			var entry CacheEntry
			if err := json.Unmarshal([]byte(data), &entry); err == nil {
				// Check if entry is still valid
				if time.Since(entry.Timestamp) < entry.TTL {
					return entry.Data, true
				}
				// Remove expired entry
				cs.redisClient.Del(ctx, key)
			}
		}
	}

	// Fallback to memory cache
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	if entry, exists := cs.memoryCache[key]; exists {
		// Check if entry is still valid
		if time.Since(entry.Timestamp) < entry.TTL {
			return entry.Data, true
		}
		// Remove expired entry
		delete(cs.memoryCache, key)
	}

	return nil, false
}

// Delete deletes a value from cache
func (cs *CacheService) Delete(ctx context.Context, key string) error {
	if !cs.enabled {
		return nil
	}

	// Delete from Redis
	if cs.redisClient != nil {
		cs.redisClient.Del(ctx, key)
	}

	// Delete from memory cache
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	delete(cs.memoryCache, key)

	return nil
}

// Clear clears all cache
func (cs *CacheService) Clear(ctx context.Context) error {
	if !cs.enabled {
		return nil
	}

	// Clear Redis cache (this is a simplified approach)
	if cs.redisClient != nil {
		// Note: In production, you might want to use a more specific pattern
		// or maintain a list of cache keys for selective clearing
		cs.redisClient.FlushDB(ctx)
	}

	// Clear memory cache
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	cs.memoryCache = make(map[string]CacheEntry)

	return nil
}

// CleanupExpired removes expired entries from memory cache
func (cs *CacheService) CleanupExpired() {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	now := time.Now()
	for key, entry := range cs.memoryCache {
		if now.Sub(entry.Timestamp) >= entry.TTL {
			delete(cs.memoryCache, key)
		}
	}
}

// CacheMiddleware provides HTTP response caching
func CacheMiddleware(cacheService *CacheService, ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip caching for non-GET requests
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		// Skip caching for authenticated requests that shouldn't be cached
		if c.GetHeader("Cache-Control") == "no-cache" {
			c.Next()
			return
		}

		key := cacheService.generateKey(c)

		// Try to get from cache
		if cached, found := cacheService.Get(c.Request.Context(), key); found {
			logger := logging.GetLoggerFromContext(c)
			logger.Debug("Cache hit",
				zap.String("key", key),
				zap.String("path", c.Request.URL.Path),
			)

			c.JSON(200, cached)
			c.Abort()
			return
		}

		// Cache miss, continue with request
		logger := logging.GetLoggerFromContext(c)
		logger.Debug("Cache miss",
			zap.String("key", key),
			zap.String("path", c.Request.URL.Path),
		)

		// Store original response writer
		originalWriter := c.Writer

		// Create a custom response writer to capture the response
		responseWriter := &responseCapture{
			ResponseWriter: originalWriter,
			body:           make([]byte, 0),
		}
		c.Writer = responseWriter

		c.Next()

		// Cache successful responses
		if responseWriter.Status() == 200 && len(responseWriter.body) > 0 {
			var responseData interface{}
			if err := json.Unmarshal(responseWriter.body, &responseData); err == nil {
				cacheService.Set(c.Request.Context(), key, responseData, ttl)
			}
		}
	}
}

// responseCapture captures the response for caching
type responseCapture struct {
	gin.ResponseWriter
	body   []byte
	status int
}

func (rc *responseCapture) Write(b []byte) (int, error) {
	rc.body = append(rc.body, b...)
	return rc.ResponseWriter.Write(b)
}

func (rc *responseCapture) WriteString(s string) (int, error) {
	rc.body = append(rc.body, []byte(s)...)
	return rc.ResponseWriter.WriteString(s)
}

func (rc *responseCapture) WriteHeader(code int) {
	rc.status = code
	rc.ResponseWriter.WriteHeader(code)
}

func (rc *responseCapture) Status() int {
	return rc.status
}

// StartCleanup starts periodic cleanup of expired cache entries
func (cs *CacheService) StartCleanup(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			cs.CleanupExpired()
		}
	}()
}

// GetStats returns cache statistics
func (cs *CacheService) GetStats() map[string]interface{} {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	return map[string]interface{}{
		"enabled":        cs.enabled,
		"memory_entries": len(cs.memoryCache),
		"redis_enabled":  cs.redisClient != nil,
	}
}
