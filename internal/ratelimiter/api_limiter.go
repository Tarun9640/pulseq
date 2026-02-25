package ratelimiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type APIRateLimiter struct {
	RedisClient *redis.Client
	Limit       int
}

func NewAPIRateLimiter(client *redis.Client, limit int) *APIRateLimiter {
	return &APIRateLimiter{
		RedisClient: client,
		Limit: limit,
	}
} 

func (r *APIRateLimiter) Allow(ctx context.Context, ip string) (bool, error) {

	// Fixed window per minute
	key := fmt.Sprintf("api_rate:%s:%d", ip, time.Now().Unix()/60)

	count, err := r.RedisClient.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	// First request ki expiry set chestham
	if count == 1 {
		r.RedisClient.Expire(ctx, key, time.Minute)
	}

	if count > int64(r.Limit) {
		return false, nil
	}

	return true, nil
}
