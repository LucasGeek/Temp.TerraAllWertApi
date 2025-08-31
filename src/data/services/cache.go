package services

import (
	"github.com/redis/go-redis/v9"
)

type CacheService struct {
	redis *redis.Client
}

func NewCacheService(redis *redis.Client) *CacheService {
	return &CacheService{
		redis: redis,
	}
}

// Add cache methods here as needed
// Example:
// func (s *CacheService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
//     return s.redis.Set(ctx, key, value, expiration).Err()
// }