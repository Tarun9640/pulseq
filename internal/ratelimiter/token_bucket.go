package ratelimiter

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokenBucketLimiter struct {
	RedisClient *redis.Client
	Rate int        //tokens per minutes(usually in sec but for testing)
	Burst int		// bucket size
	key string 		// bucket key
}

// constructor
func NewTokenBucketLimiter(client *redis.Client, rate int, burst int, key string) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		RedisClient: client,
		Rate: rate,
		Burst: burst,
		key: key,
	}
}

// using txpipeline
// func (t *TokenBucketLimiter) Allow(ctx context.Context) (bool, error) {

// 	tokenKey := t.key + ":tokens"
// 	timeKey := t.key + ":last"

// 	now := time.Now().Unix()
// 	pipe := t.RedisClient.TxPipeline()

// 	tokencmd := pipe.Get(ctx, tokenKey)
// 	timecmd := pipe.Get(ctx, timeKey)

// 	_, err := pipe.Exec(ctx)

// 	if err != nil {
// 		log.Println("Redis pipeline error:", err)
// 	}
	

// 	var tokens float64
// 	var last int64

// 	//First Time Case -- tokens = burst
// 	if tokencmd.Val() == "" {
// 		tokens = float64(t.Burst)
// 		last = now
// 	} else {
// 		tokens, _ = tokencmd.Float64()
// 		last, _ = timecmd.Int64()
// 	}
// 	//refill tokens

// 	elapsed := now-last

// 	tokens += float64(elapsed)/60.0 * float64(t.Rate)

// 	//Burst Limit Apply tokens = 9 burst = 5 we made tokens = 5
// 	if tokens > float64(t.Burst){
// 		tokens = float64(t.Burst)
// 	}

// 	// Not enough tokens
// 	if tokens < 1 {
// 		return false,nil
// 	}

// 	// Consume token
// 	tokens -=1

// 	// Save state
// 	t.RedisClient.Set(ctx, tokenKey, tokens, 0)
// 	t.RedisClient.Set(ctx, timeKey, now, 0)

// 	return true, nil
// }  


var tokenBucketScript = redis.NewScript(`
	local tokenKey = KEYS[1]
	local timeKey = KEYS[2]

	local rate = tonumber(ARGV[1])
	local burst = tonumber(ARGV[2])
	local now = tonumber(ARGV[3])

	local tokens = tonumber(redis.call("GET", tokenKey))
	local last = tonumber(redis.call("GET", timeKey))

	-- First time initialize
	if tokens == nil then
		tokens = burst
		last = now
	end

	-- Calculate elapsed time
	local elapsed = now - last

	-- Refill tokens
	tokens = tokens + (elapsed/60)*rate

	-- Apply burst limit
	if tokens > burst then
		tokens = burst
	end

	-- Reject if no tokens
	if tokens < 1 then
		return 0
	end

	-- Consume token
	tokens = tokens - 1

	-- Save state
	redis.call("SET", tokenKey, tokens)
	redis.call("SET", timeKey, now)

	return 1
`)

func (t *TokenBucketLimiter) Allow(ctx context.Context) (bool, error) {
	tokenKey := t.key + ":tokens"
	timeKey := t.key + ":last"

	now := time.Now().Unix()
	
	result, err := tokenBucketScript.Run(ctx, t.RedisClient, []string{tokenKey, timeKey}, t.Rate, t.Burst, now).Int()

	if err != nil {
		return false, err
	}

	return result == 1,nil
}  