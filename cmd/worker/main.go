// Multi-binary architecture.
//responsibilites
// - start workers
// - load redis
// - wire queries

package main

import (
	"log"

	"github.com/Tarun9640/pulseq/internal/db"
	"github.com/Tarun9640/pulseq/pkg/postgres"
	"github.com/Tarun9640/pulseq/pkg/redis"
	"github.com/Tarun9640/pulseq/internal/worker"
)

func main() {
	log.Println("Worker started...")

	//db
	pool := postgres.NewPool()
	queries := db.New(pool)

	redisClient := redis.NewClient()

	workerCount := 5

	for i := 1; i <= workerCount; i++ {
		go worker.StartWorker(i, queries, redisClient)
		log.Printf("worker %d started...",i)
	}

	select{} // This blocks forever.If main exits → all goroutines die.
}

