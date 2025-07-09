package ratelimit

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kart2405/API_Gateway/internal/config"
	"github.com/kart2405/API_Gateway/internal/services"
	"github.com/redis/go-redis/v9"
)

const (
	DEFAULT_MAX_REQUESTS = 10
	DEFAULT_WINDOW       = 60 * time.Second
)

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDRaw, exists := c.Get("userID")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized - missing userID in context"})
			return
		}
		userID := fmt.Sprintf("%v", userIDRaw)

		if userID == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing X-User-ID header"})
			return
		}

		// Get service name from path to determine rate limit
		serviceName := c.Param("service")
		var maxRequests int
		var window time.Duration

		if serviceName != "" {
			// Try to get route-specific rate limit
			if route, exists := services.GlobalRouteOptimizer.FindRouteOptimized(serviceName); exists {
				maxRequests = route.RateLimit
				window = time.Duration(route.RateLimitWindow) * time.Second
			}
		}

		// Use defaults if no route-specific config found
		if maxRequests == 0 {
			maxRequests = DEFAULT_MAX_REQUESTS
		}
		if window == 0 {
			window = DEFAULT_WINDOW
		}

		key := fmt.Sprintf("ratelimit:%s:%s", userID, serviceName)

		script := redis.NewScript(`
			local tokens = redis.call("GET", KEYS[1])
			if not tokens then
				redis.call("SETEX", KEYS[1], ARGV[2], ARGV[1])
				return ARGV[1]
			end
			if tonumber(tokens) > 0 then
				return redis.call("DECR", KEYS[1])
			else
				return -1
			end
		`)

		result, err := script.Run(config.Ctx, config.RedisClient, []string{key}, maxRequests, int(window.Seconds())).Result()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Redis error"})
			return
		}

		tokensLeft, _ := strconv.Atoi(fmt.Sprintf("%v", result))
		if tokensLeft < 0 {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			return
		}

		c.Next()
	}
}
