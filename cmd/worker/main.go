// Multi-binary architecture.
//responsibilites
// - start workers
// - load redis
// - wire queries

package main

import (
	"github.com/Tarun9640/pulseq/internal/db"
	"github.com/Tarun9640/pulseq/pkg/postgres"
	"github.com/Tarun9640/pulseq/pkg/redis"
	"github.com/Tarun9640/pulseq/internal/worker"
)

func main() {

	//db
	pool := postgres.NewPool()
	queries := db.New(pool)

	redisClient := redis.NewClient()

	go worker.Scheduler(redisClient)

	workerCount := 5

	for i := 1; i <= workerCount; i++ {
		go worker.StartWorker(i, queries, redisClient)
	}

	select{} // This blocks forever.If main exits → all goroutines die.
}

