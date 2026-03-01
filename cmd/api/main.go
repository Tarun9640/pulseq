package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/Tarun9640/pulseq/internal/config"
	"github.com/Tarun9640/pulseq/internal/db"
	"github.com/Tarun9640/pulseq/internal/handler"
	"github.com/Tarun9640/pulseq/internal/middleware"
	"github.com/Tarun9640/pulseq/internal/queue"
	"github.com/Tarun9640/pulseq/internal/ratelimiter"
	"github.com/Tarun9640/pulseq/internal/repository"
	"github.com/Tarun9640/pulseq/internal/service"
	"github.com/Tarun9640/pulseq/pkg/postgres"
	"github.com/Tarun9640/pulseq/pkg/redis"
	"github.com/gin-gonic/gin"
)


func main() {

	cfg := config.LoadConfig()

	// create pool
	pool := postgres.NewPool(cfg)

	// create sqlc Queries
	queries := db.New(pool)

	log.Println("sqlc connected...")

	//redis
	redisClient := redis.NewClient(cfg)
	taskQueue := queue.NewRedisQueue(redisClient)

	//layers
	repo := repository.NewTaskRepository(queries)
	//inject -happens here
	svc := service.NewTaskService(repo, taskQueue)
	handler := handler.NewTaskHandler(svc)

	//gin
	router := gin.Default()

	// API Rate Limiter
	apiRateLimiter := ratelimiter.NewAPIRateLimiter(redisClient, cfg.APIRateLimit)

	// Health API (No rate limit)
	router.GET("/health", func(c *gin.Context){
		c.JSON(200,"OK")
	})

	taskGroup := router.Group("/tasks")

	taskGroup.Use(
		middleware.RateLimitMiddleware(apiRateLimiter),
	)

	taskGroup.POST("", handler.CreateTask)
	//router.Run(":8080")

	// HTTP Server
	srv := &http.Server{
		Addr: ":" + cfg.Port,
		Handler: router,
	}

	// Start server
	go func() {
		log.Println("Server running on :8080")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server error:", err)
		}
	}()

	// Shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	<-quit

	log.Println("Shutdown signal received")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("API server stopped gracefully")

}