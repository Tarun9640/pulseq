package worker

import (
	"context"
	"log"
	"strconv"
	"time"
	"github.com/Tarun9640/pulseq/internal/queue"
	redisLib "github.com/redis/go-redis/v9"
)

func Scheduler(ctx context.Context, redisClient *redisLib.Client) {

	log.Println("Scheduler started...")

	for {

		select {

		case <-ctx.Done():
			log.Println("Graceful shutdown complete")
			return

		default:
			now := time.Now().Unix()

			tasks, err := redisClient.ZRangeByScore(ctx, queue.DelayQueue, &redisLib.ZRangeBy{
				Min:    "-inf",
				Max:    strconv.FormatInt(now, 10),
				Offset: 0,
				Count:  10,
			}).Result()

			if err != nil {
				log.Println("Scheduler error:", err)
				time.Sleep(time.Second)
				continue
			}

			for _, taskID := range tasks {
				// Move atomically (best practice is Lua, but keeping simple) -- redis provide lua
				// While Lua script runs: No other command runs , No interruption, No partial state
				redisClient.ZRem(ctx, queue.DelayQueue, taskID)
				redisClient.LPush(ctx, queue.TaskQueueName, taskID)

				log.Printf("scheduler moved task %s to main queue",taskID)
			}

			time.Sleep(1 * time.Second)
		}	
	}
}