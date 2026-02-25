package ratelimiter

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRateLimiter struct {
	RedisClient *redis.Client
	Limit int
}

func NewRedisRateLimiter(redisClient *redis.Client, limit int) *RedisRateLimiter {
	return &RedisRateLimiter{
		RedisClient: redisClient,
		Limit: limit,
	}
}

//Fixed Window
func (r *RedisRateLimiter) Allow(ctx context.Context) (bool, error) {

	//key per second
	key := fmt.Sprintf("Rate limit : %d", time.Now().Unix()/60)

	count, err := r.RedisClient.Incr(ctx, key).Result()

	if err != nil {
		log.Println("Error in increment counter")
		return false, err
	}

	// Set expiry only when first increment happens
	if count == 1 {
		err = r.RedisClient.Expire(ctx, key, time.Minute).Err()
		if err != nil {
			return false, err
		}
	}

	if count > int64(r.Limit) {
		return false, nil
	}

	return true, nil
}