package redis

import (
	"context"
	"log"

	"github.com/Tarun9640/pulseq/internal/config"
	"github.com/redis/go-redis/v9"
)

func NewClient(cfg config.Config) *redis.Client {

	client := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})

	_, err := client.Ping(context.Background()).Result()

	if err != nil {
		log.Fatal("Redis not responding:", err)
	}

	log.Println("Connected to Redis")

	return client
}
