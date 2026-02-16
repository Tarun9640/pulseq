package main

import (
	"log"

	"github.com/Tarun9640/pulseq/internal/db"
	"github.com/Tarun9640/pulseq/internal/handler"
	"github.com/Tarun9640/pulseq/internal/queue"
	"github.com/Tarun9640/pulseq/internal/repository"
	"github.com/Tarun9640/pulseq/internal/service"
	"github.com/Tarun9640/pulseq/pkg/postgres"
	"github.com/Tarun9640/pulseq/pkg/redis"
	"github.com/gin-gonic/gin"
)


func main() {
	// create pool
	pool := postgres.NewPool()

	// create sqlc Queries
	queries := db.New(pool)

	log.Println("sqlc connected...")

	//redis
	redisClient := redis.NewClient()
	taskQueue := queue.NewRedisQueue(redisClient)

	//layers
	repo := repository.NewTaskRepository(queries)
	//inject -happens here
	svc := service.NewTaskService(repo, taskQueue)
	handler := handler.NewTaskHandler(svc)

	//gin
	router := gin.Default()

	router.POST("/tasks", handler.CreateTask)

	log.Println("Server running on :8080")

	router.Run(":8080")
}