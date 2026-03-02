// Multi-binary architecture.
//responsibilites
// - start workers
// - load redis
// - wire queries

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Tarun9640/pulseq/internal/config"
	"github.com/Tarun9640/pulseq/internal/db"
	"github.com/Tarun9640/pulseq/internal/handler"
	"github.com/Tarun9640/pulseq/internal/logger"
	"github.com/Tarun9640/pulseq/internal/ratelimiter"
	"github.com/Tarun9640/pulseq/internal/worker"
	"github.com/Tarun9640/pulseq/pkg/postgres"
	"github.com/Tarun9640/pulseq/pkg/redis"
	"github.com/gin-gonic/gin"
)

func main() {

	cfg := config.LoadConfig()

	//db
	pool := postgres.NewPool(cfg)
	queries := db.New(pool)


	redisClient := redis.NewClient(cfg)

	ctx, cancel := context.WithCancel(context.Background())

	go worker.Scheduler(ctx, redisClient)

	logger := logger.NewLogger()

	// limiter := rate.NewLimiter(rate.Limit(config.RatelimitRPS), config.RatelimitBurst)

	//redisLimiter := ratelimiter.NewRedisRateLimiter(redisClient, 3) // only 3 workers are allowed per sec

	//System allows 3 immediate requests burst, then 1 request per minute refill
	tokenLimiter := ratelimiter.NewTokenBucketLimiter(redisClient, cfg.TokenRate, cfg.TokenBurst, "worker_bucket")

	// workerCount := 5

	// for i := 1; i <= workerCount; i++ {
	// 	go worker.StartWorker(i, queries, redisClient, tokenLimiter)
	// }

	// Dynamic Worker Manager
	manager := worker.NewWorkerManager(redisClient, tokenLimiter, queries, cfg.WorkerMin, cfg.WorkerMax, logger)

	go manager.Start(ctx)


	//metrics server
	metricsHandler := handler.NewMetricsHandler(redisClient, manager, logger)

	workerHandler := handler.NewWorkerHandler(manager)

	router := gin.Default()

	router.GET("/metrics", metricsHandler.GetMetrics)

	router.GET("/workers", workerHandler.GetWorkers)

	// workers run on 9090 
	go router.Run(":9090")


	//select{} // This blocks forever.If main exits → all goroutines die.

	//signal handling for graceful shutdown
	stopSignal := make(chan os.Signal, 1)

	signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM)

	<-stopSignal

	logger.Info("Shutdown signal received")

	cancel() // stop scheduler + manager

	manager.Stop()

	logger.Info("Graceful shutdown complete")
}

