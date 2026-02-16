package redis

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

func NewClient() *redis.Client {

	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6380",
	})

	_, err := client.Ping(context.Background()).Result()

	if err != nil {
		log.Fatal("Redis not responding:", err)
	}

	log.Println("Connected to Redis")

	return client
}
