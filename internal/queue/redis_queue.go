package queue

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type RedisQueue struct {
	client *redis.Client
}

func NewRedisQueue(c *redis.Client) *RedisQueue {
	return &RedisQueue{
		client: c,
	}
}

const (
	TaskQueueName = "tasks_queue"
	ProcessingQueueName = "processing_queue"
	DeadLetterQueue = "tasks_dlq"
	DelayQueue = "delay_queue"
)

//LPUSH → add job
//BRPOP → worker takes job


func (q *RedisQueue) Enqueue(ctx context.Context, taskID string) error {

	return q.client.LPush(ctx, TaskQueueName, taskID).Err()
}


