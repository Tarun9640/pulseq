// Multi-binary architecture.
//responsibilites
// - start workers
// - load redis
// - wire queries

package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/Tarun9640/pulseq/internal/db"
	"github.com/Tarun9640/pulseq/internal/ratelimiter"
	"github.com/Tarun9640/pulseq/internal/worker"
	"github.com/Tarun9640/pulseq/pkg/postgres"
	"github.com/Tarun9640/pulseq/pkg/redis"
)

func main() {

	//db
	pool := postgres.NewPool()
	queries := db.New(pool)

	redisClient := redis.NewClient()

	go worker.Scheduler(redisClient)

	// config := config.Config{
	// 	RatelimitRPS: 1, //5 tokens per second
	// 	RatelimitBurst: 1, //bucketsize
	// }

	// limiter := rate.NewLimiter(rate.Limit(config.RatelimitRPS), config.RatelimitBurst)

	//redisLimiter := ratelimiter.NewRedisRateLimiter(redisClient, 3) // only 3 workers are allowed per sec

	//System allows 3 immediate requests burst, then 1 request per minute refill
	tokenLimiter := ratelimiter.NewTokenBucketLimiter(redisClient, 100, 50, "worker_bucket")

	// workerCount := 5

	// for i := 1; i <= workerCount; i++ {
	// 	go worker.StartWorker(i, queries, redisClient, tokenLimiter)
	// }

	// Dynamic Worker Manager
	manager := worker.NewWorkerManager(redisClient, tokenLimiter, queries, 1, 10)

	go manager.Start()

	//select{} // This blocks forever.If main exits → all goroutines die.

	//signal handling for graceful shutdown
	stopSignal := make(chan os.Signal, 1)

	signal.Notify(stopSignal, os.Interrupt)

	<-stopSignal

	log.Println("Shutdown signal received")

	manager.Stop()

}

