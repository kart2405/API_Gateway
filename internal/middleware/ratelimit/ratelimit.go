package ratelimit

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kart2405/API_Gateway/internal/config"
	"github.com/redis/go-redis/v9"
)

const (
	MAX_REQUESTS = 10
	WINDOW       = 60 * time.Second
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

		key := fmt.Sprintf("ratelimit:%s", userID)

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

		result, err := script.Run(config.Ctx, config.RedisClient, []string{key}, MAX_REQUESTS, int(WINDOW.Seconds())).Result()
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
