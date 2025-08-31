package middleware

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

type RateLimiterConfig struct {
	RequestsPerMinute int
	WindowSize        time.Duration
	RedisClient       *redis.Client
}

func NewRateLimiter(config RateLimiterConfig) fiber.Handler {
	if config.RequestsPerMinute == 0 {
		config.RequestsPerMinute = 100 // default
	}
	if config.WindowSize == 0 {
		config.WindowSize = time.Minute // default
	}

	return func(c *fiber.Ctx) error {
		// Get client identifier (IP or user ID if authenticated)
		clientID := getClientID(c)
		
		key := fmt.Sprintf("rate_limit:%s", clientID)
		ctx := c.Context()

		// Get current count
		count, err := config.RedisClient.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			// If Redis is down, allow request but log error
			return c.Next()
		}

		if count >= config.RequestsPerMinute {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Rate limit exceeded",
				"retry_after": config.WindowSize.Seconds(),
			})
		}

		// Increment counter
		pipe := config.RedisClient.Pipeline()
		pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, config.WindowSize)
		_, err = pipe.Exec(ctx)
		if err != nil {
			// Log error but continue
		}

		// Add rate limit headers
		c.Set("X-RateLimit-Limit", strconv.Itoa(config.RequestsPerMinute))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(config.RequestsPerMinute-count-1))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(config.WindowSize).Unix(), 10))

		return c.Next()
	}
}

func getClientID(c *fiber.Ctx) string {
	// Try to get user ID from JWT claims first
	if user := c.Locals("user"); user != nil {
		if claims, ok := user.(map[string]interface{}); ok {
			if userID, exists := claims["user_id"].(string); exists {
				return fmt.Sprintf("user:%s", userID)
			}
		}
	}
	
	// Fall back to IP address
	return fmt.Sprintf("ip:%s", c.IP())
}