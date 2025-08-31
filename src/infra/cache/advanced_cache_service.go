package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"api/infra/logger"
)

type CacheConfig struct {
	DefaultTTL      time.Duration
	ShortTTL        time.Duration
	LongTTL         time.Duration
	MaxRetries      int
	RetryInterval   time.Duration
}

type AdvancedCacheService struct {
	redisClient *redis.Client
	config      CacheConfig
}

type CacheKey string

const (
	// Cache key patterns
	TowersListKey          CacheKey = "towers:list"
	TowerByIDKey          CacheKey = "tower:id:%s"
	ApartmentsByFloorKey  CacheKey = "apartments:floor:%s"
	ApartmentByIDKey      CacheKey = "apartment:id:%s"
	ApartmentSearchKey    CacheKey = "apartments:search:%s"
	UserProfileKey        CacheKey = "user:profile:%s"
	GalleryImagesKey      CacheKey = "gallery:images:%s"
)

func NewAdvancedCacheService(redisClient *redis.Client) *AdvancedCacheService {
	return &AdvancedCacheService{
		redisClient: redisClient,
		config: CacheConfig{
			DefaultTTL:    15 * time.Minute,
			ShortTTL:      5 * time.Minute,
			LongTTL:       2 * time.Hour,
			MaxRetries:    3,
			RetryInterval: 100 * time.Millisecond,
		},
	}
}

// Set caches a value with optimized TTL based on data type
func (c *AdvancedCacheService) Set(ctx context.Context, key CacheKey, value interface{}, ttl ...time.Duration) error {
	keyStr := string(key)
	
	// Determine optimal TTL
	cacheTTL := c.getOptimalTTL(key, ttl...)
	
	// Serialize value
	data, err := json.Marshal(value)
	if err != nil {
		logger.Error(ctx, "Failed to marshal cache value", err, zap.String("key", keyStr))
		return err
	}
	
	// Set with retry mechanism
	for i := 0; i < c.config.MaxRetries; i++ {
		err = c.redisClient.Set(ctx, keyStr, data, cacheTTL).Err()
		if err == nil {
			logger.Debug(ctx, "Cache set successfully", 
				zap.String("key", keyStr),
				zap.Duration("ttl", cacheTTL),
			)
			return nil
		}
		
		if i < c.config.MaxRetries-1 {
			time.Sleep(c.config.RetryInterval)
		}
	}
	
	logger.Error(ctx, "Failed to set cache after retries", err, zap.String("key", keyStr))
	return err
}

// Get retrieves and deserializes a cached value
func (c *AdvancedCacheService) Get(ctx context.Context, key CacheKey, dest interface{}) error {
	keyStr := string(key)
	
	// Get with retry mechanism
	var data []byte
	var err error
	
	for i := 0; i < c.config.MaxRetries; i++ {
		result := c.redisClient.Get(ctx, keyStr)
		data, err = result.Bytes()
		
		if err == nil {
			break
		}
		
		if err == redis.Nil {
			return ErrCacheMiss
		}
		
		if i < c.config.MaxRetries-1 {
			time.Sleep(c.config.RetryInterval)
		}
	}
	
	if err != nil {
		logger.Error(ctx, "Failed to get cache after retries", err, zap.String("key", keyStr))
		return err
	}
	
	// Deserialize
	err = json.Unmarshal(data, dest)
	if err != nil {
		logger.Error(ctx, "Failed to unmarshal cache value", err, zap.String("key", keyStr))
		return err
	}
	
	logger.Debug(ctx, "Cache hit", zap.String("key", keyStr))
	return nil
}

// GetOrSet implements cache-aside pattern with loader function
func (c *AdvancedCacheService) GetOrSet(ctx context.Context, key CacheKey, dest interface{}, loader func() (interface{}, error), ttl ...time.Duration) error {
	// Try to get from cache first
	err := c.Get(ctx, key, dest)
	if err == nil {
		return nil // Cache hit
	}
	
	if err != ErrCacheMiss {
		// Actual error, log and continue with loader
		logger.Warn(ctx, "Cache get error, falling back to loader", zap.Error(err))
	}
	
	// Cache miss, use loader
	value, err := loader()
	if err != nil {
		return err
	}
	
	// Set in cache (fire and forget)
	go func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		if setErr := c.Set(cacheCtx, key, value, ttl...); setErr != nil {
			logger.Warn(cacheCtx, "Failed to set cache in background", zap.Error(setErr))
		}
	}()
	
	// Copy value to destination
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	
	return json.Unmarshal(data, dest)
}

// Delete removes a key from cache
func (c *AdvancedCacheService) Delete(ctx context.Context, key CacheKey) error {
	keyStr := string(key)
	
	err := c.redisClient.Del(ctx, keyStr).Err()
	if err != nil {
		logger.Error(ctx, "Failed to delete cache key", err, zap.String("key", keyStr))
		return err
	}
	
	logger.Debug(ctx, "Cache key deleted", zap.String("key", keyStr))
	return nil
}

// DeletePattern removes all keys matching a pattern
func (c *AdvancedCacheService) DeletePattern(ctx context.Context, pattern string) error {
	keys, err := c.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		logger.Error(ctx, "Failed to get keys for pattern", err, zap.String("pattern", pattern))
		return err
	}
	
	if len(keys) == 0 {
		return nil
	}
	
	err = c.redisClient.Del(ctx, keys...).Err()
	if err != nil {
		logger.Error(ctx, "Failed to delete cache keys by pattern", err, zap.String("pattern", pattern))
		return err
	}
	
	logger.Debug(ctx, "Cache keys deleted by pattern", 
		zap.String("pattern", pattern),
		zap.Int("count", len(keys)),
	)
	return nil
}

// Exists checks if a key exists in cache
func (c *AdvancedCacheService) Exists(ctx context.Context, key CacheKey) (bool, error) {
	result := c.redisClient.Exists(ctx, string(key))
	count, err := result.Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetTTL returns the remaining TTL for a key
func (c *AdvancedCacheService) GetTTL(ctx context.Context, key CacheKey) (time.Duration, error) {
	return c.redisClient.TTL(ctx, string(key)).Result()
}

// ExtendTTL extends the TTL of an existing key
func (c *AdvancedCacheService) ExtendTTL(ctx context.Context, key CacheKey, extension time.Duration) error {
	return c.redisClient.Expire(ctx, string(key), extension).Err()
}

// getOptimalTTL determines the best TTL based on key type and usage patterns
func (c *AdvancedCacheService) getOptimalTTL(key CacheKey, customTTL ...time.Duration) time.Duration {
	if len(customTTL) > 0 {
		return customTTL[0]
	}
	
	keyStr := string(key)
	
	// Static/reference data - longer TTL
	if keyStr == string(TowersListKey) {
		return c.config.LongTTL
	}
	
	// User-specific data - shorter TTL
	if keyStr == string(UserProfileKey) {
		return c.config.ShortTTL
	}
	
	// Search results - very short TTL
	if keyStr == string(ApartmentSearchKey) {
		return c.config.ShortTTL
	}
	
	// Default TTL
	return c.config.DefaultTTL
}

// Cache invalidation helpers
func (c *AdvancedCacheService) InvalidateTowerCache(ctx context.Context, towerID string) error {
	patterns := []string{
		string(TowersListKey),
		fmt.Sprintf(string(TowerByIDKey), towerID),
		fmt.Sprintf("apartments:tower:%s*", towerID),
	}
	
	for _, pattern := range patterns {
		if err := c.DeletePattern(ctx, pattern); err != nil {
			return err
		}
	}
	
	return nil
}

func (c *AdvancedCacheService) InvalidateApartmentCache(ctx context.Context, apartmentID, floorID string) error {
	patterns := []string{
		fmt.Sprintf(string(ApartmentByIDKey), apartmentID),
		fmt.Sprintf(string(ApartmentsByFloorKey), floorID),
		string(ApartmentSearchKey) + "*",
	}
	
	for _, pattern := range patterns {
		if err := c.DeletePattern(ctx, pattern); err != nil {
			return err
		}
	}
	
	return nil
}

// Health check
func (c *AdvancedCacheService) HealthCheck(ctx context.Context) error {
	return c.redisClient.Ping(ctx).Err()
}

// Cache statistics
type CacheStats struct {
	HitRate    float64 `json:"hitRate"`
	TotalKeys  int64   `json:"totalKeys"`
	UsedMemory string  `json:"usedMemory"`
	Uptime     int64   `json:"uptime"`
}

func (c *AdvancedCacheService) GetStats(ctx context.Context) (*CacheStats, error) {
	keyCount := c.redisClient.DBSize(ctx)
	totalKeys, err := keyCount.Result()
	if err != nil {
		return nil, err
	}
	
	// Parse Redis INFO output (simplified)
	stats := &CacheStats{
		TotalKeys: totalKeys,
		HitRate:   0.0, // Would need to parse from INFO stats
	}
	
	logger.Debug(ctx, "Cache stats", zap.Any("stats", stats))
	return stats, nil
}

// Error definitions
var (
	ErrCacheMiss = fmt.Errorf("cache miss")
)