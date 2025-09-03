package middleware

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/time/rate"
)

// FileSizeCategory represents different file size categories with different limits
type FileSizeCategory string

const (
	SmallFile  FileSizeCategory = "small"  // < 1MB
	MediumFile FileSizeCategory = "medium" // 1MB - 100MB
	LargeFile  FileSizeCategory = "large"  // > 100MB
)

// UploadRateLimiter implements file size-based rate limiting
type UploadRateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	config   RateLimitConfig
}

type RateLimitConfig struct {
	SmallFileRate  rate.Limit    // requests per minute for files < 1MB
	MediumFileRate rate.Limit    // requests per minute for files 1MB-100MB
	LargeFileRate  rate.Limit    // requests per minute for files > 100MB
	BurstSize      int           // burst capacity
	CleanupTicker  time.Duration // cleanup interval for expired limiters
}

// DefaultRateLimitConfig returns production-ready rate limiting configuration
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		SmallFileRate:  rate.Every(time.Minute / 50), // 50 uploads per minute
		MediumFileRate: rate.Every(time.Minute / 5),  // 5 uploads per minute
		LargeFileRate:  rate.Every(10 * time.Minute), // 1 upload per 10 minutes
		BurstSize:      1,
		CleanupTicker:  5 * time.Minute,
	}
}

// NewUploadRateLimiter creates a new file size-based rate limiter
func NewUploadRateLimiter(config RateLimitConfig) *UploadRateLimiter {
	limiter := &UploadRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		config:   config,
	}

	// Start cleanup goroutine to remove inactive limiters
	go limiter.cleanup()

	return limiter
}

// GetFileSizeCategory determines the category based on file size
func GetFileSizeCategory(fileSize int64) FileSizeCategory {
	switch {
	case fileSize < 1*1024*1024: // < 1MB
		return SmallFile
	case fileSize < 100*1024*1024: // < 100MB
		return MediumFile
	default: // >= 100MB
		return LargeFile
	}
}

// Allow checks if an upload is allowed for the given user and file size
func (url *UploadRateLimiter) Allow(userID string, fileSize int64) bool {
	url.mu.Lock()
	defer url.mu.Unlock()

	category := GetFileSizeCategory(fileSize)
	key := fmt.Sprintf("%s:%s", userID, category)

	limiter, exists := url.limiters[key]
	if !exists {
		var rateLimit rate.Limit
		switch category {
		case SmallFile:
			rateLimit = url.config.SmallFileRate
		case MediumFile:
			rateLimit = url.config.MediumFileRate
		case LargeFile:
			rateLimit = url.config.LargeFileRate
		}

		limiter = rate.NewLimiter(rateLimit, url.config.BurstSize)
		url.limiters[key] = limiter
	}

	return limiter.Allow()
}

// ReserveN reserves n tokens for future use
func (url *UploadRateLimiter) ReserveN(userID string, fileSize int64, n int) *rate.Reservation {
	url.mu.Lock()
	defer url.mu.Unlock()

	category := GetFileSizeCategory(fileSize)
	key := fmt.Sprintf("%s:%s", userID, category)

	limiter, exists := url.limiters[key]
	if !exists {
		var rateLimit rate.Limit
		switch category {
		case SmallFile:
			rateLimit = url.config.SmallFileRate
		case MediumFile:
			rateLimit = url.config.MediumFileRate
		case LargeFile:
			rateLimit = url.config.LargeFileRate
		}

		limiter = rate.NewLimiter(rateLimit, url.config.BurstSize)
		url.limiters[key] = limiter
	}

	return limiter.ReserveN(time.Now(), n)
}

// GetRateLimit returns the current rate limit for a user and file size
func (url *UploadRateLimiter) GetRateLimit(userID string, fileSize int64) rate.Limit {
	category := GetFileSizeCategory(fileSize)
	switch category {
	case SmallFile:
		return url.config.SmallFileRate
	case MediumFile:
		return url.config.MediumFileRate
	case LargeFile:
		return url.config.LargeFileRate
	}
	return url.config.LargeFileRate // Default to most restrictive
}

// cleanup removes inactive rate limiters periodically
func (url *UploadRateLimiter) cleanup() {
	ticker := time.NewTicker(url.config.CleanupTicker)
	defer ticker.Stop()

	for range ticker.C {
		url.mu.Lock()
		// Remove limiters that haven't been used recently
		// This prevents memory leaks from inactive users
		for key, limiter := range url.limiters {
			// Check if limiter has available tokens (indicating no recent usage)
			if limiter.TokensAt(time.Now()) >= float64(url.config.BurstSize) {
				delete(url.limiters, key)
			}
		}
		url.mu.Unlock()
	}
}

// UploadRateLimitMiddleware returns a Fiber middleware for upload rate limiting
func UploadRateLimitMiddleware(rateLimiter *UploadRateLimiter) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract user ID (assuming it's set by auth middleware)
		userID := c.Locals("user_id")
		if userID == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		userIDStr, ok := userID.(string)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Invalid user ID format",
			})
		}

		// Get file size from request (assuming it's passed as header or form data)
		var fileSize int64
		
		// Try to get from Content-Length header
		if contentLength := c.Get("Content-Length"); contentLength != "" {
			if size := c.Context().Request.Header.ContentLength(); size >= 0 {
				fileSize = int64(size)
			}
		}

		// Try to get from custom header X-File-Size
		if fileSize == 0 {
			if sizeHeader := c.Get("X-File-Size"); sizeHeader != "" {
				if size, err := parseFileSize(sizeHeader); err == nil {
					fileSize = size
				}
			}
		}

		// If we still don't have file size, assume medium file for safety
		if fileSize == 0 {
			fileSize = 10 * 1024 * 1024 // 10MB default
		}

		// Check rate limit
		if !rateLimiter.Allow(userIDStr, fileSize) {
			category := GetFileSizeCategory(fileSize)
			rateLimit := rateLimiter.GetRateLimit(userIDStr, fileSize)
			
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":           "Upload rate limit exceeded",
				"file_size":       fileSize,
				"size_category":   string(category),
				"rate_limit":      fmt.Sprintf("%.2f requests per minute", float64(rateLimit)*60),
				"retry_after":     "60", // seconds
			})
		}

		return c.Next()
	}
}

// parseFileSize parses file size from string (supports units like MB, GB)
func parseFileSize(sizeStr string) (int64, error) {
	// Simple implementation - can be enhanced to parse units
	var size int64
	_, err := fmt.Sscanf(sizeStr, "%d", &size)
	return size, err
}

// CircuitBreakerConfig configuration for circuit breaker pattern
type CircuitBreakerConfig struct {
	MaxFailures     int           // Number of failures before opening circuit
	ResetTimeout    time.Duration // Time to wait before trying to close circuit
	CheckInterval   time.Duration // How often to check if circuit should reset
}

// CircuitBreaker implements the circuit breaker pattern for upload failures
type CircuitBreaker struct {
	config       CircuitBreakerConfig
	failures     int
	lastFailTime time.Time
	state        CircuitState
	mu           sync.RWMutex
}

type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
		state:  CircuitClosed,
	}
}

// IsAllowed checks if requests are allowed through the circuit breaker
func (cb *CircuitBreaker) IsAllowed() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		if time.Since(cb.lastFailTime) > cb.config.ResetTimeout {
			cb.state = CircuitHalfOpen
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	}
	return false
}

// RecordSuccess records a successful operation
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	cb.failures = 0
	cb.state = CircuitClosed
}

// RecordFailure records a failed operation
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	cb.failures++
	cb.lastFailTime = time.Now()
	
	if cb.failures >= cb.config.MaxFailures {
		cb.state = CircuitOpen
	}
}