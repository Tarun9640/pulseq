package worker

import (
	"context"
	"log"
	"time"

	"github.com/Tarun9640/pulseq/internal/db"
	"github.com/Tarun9640/pulseq/internal/queue"
	"github.com/Tarun9640/pulseq/internal/ratelimiter"
	"github.com/redis/go-redis/v9"
)

type WorkerManager struct {
	redisClient *redis.Client
	limiter *ratelimiter.TokenBucketLimiter
	queries *db.Queries

	minWorkers int
	maxWorkers int
	currentWorkers int
	stopChans map[int]chan bool  // key workerId like 1,2 Value = stop channel like stopChan1
}

func NewWorkerManager(client *redis.Client, limiter *ratelimiter.TokenBucketLimiter, queries *db.Queries, minWorkers int, maxWorkers int) *WorkerManager {
	return &WorkerManager{
		redisClient: client,
		limiter: limiter,
		queries: queries,
		minWorkers: minWorkers,
		maxWorkers: maxWorkers,
		currentWorkers: 0,
		stopChans: make(map[int]chan bool),
	}
}

func (m *WorkerManager) Start(ctx context.Context) {

	log.Println("Worker Manager started...")

	for {

		select {
			case <-ctx.Done():
				log.Println("Worker manager stopping")
				return

			default:
				// Step 1: Get current queue size from Redis
				// This tells how many tasks are waiting
				queueSize, err := m.redisClient.LLen(ctx, queue.TaskQueueName).Result()

				if err != nil {
					log.Println("Queue size error:", err)

					// Wait before retrying
					time.Sleep(5 * time.Second)
					continue
				}

				log.Println("Queue size:", queueSize)

				// Step 2: Decide how many workers needed
				// Example:
				// 0-5 tasks -> 1 worker
				// 6-10 tasks -> 2 workers
				requiredWorkers := int(queueSize/5) + 1

				// Step 3: Apply minimum workers limit
				if requiredWorkers < m.minWorkers {
					requiredWorkers = m.minWorkers
				}

				// Step 4: Apply maximum workers limit
				if requiredWorkers > m.maxWorkers {
					requiredWorkers = m.maxWorkers
				}

				log.Println("Required workers:", requiredWorkers)
				log.Println("Current workers:", m.currentWorkers)


				// Step 5: SCALE UP (Add workers)
				if requiredWorkers > m.currentWorkers {

					workersToAdd := requiredWorkers - m.currentWorkers

					log.Printf("Starting %d workers", workersToAdd)

					for i := 0; i < workersToAdd; i++ {

						// Increase worker count
						m.currentWorkers++

						// Create stop channel for this worker
						// Used to stop worker safely later
						stopChan := make(chan bool, 1)

						// Save stop channel in map
						// key = workerID
						// value = stop channel
						m.stopChans[m.currentWorkers] = stopChan

						// Start worker goroutine
						go StartWorker(
							ctx,
							m.currentWorkers,
							m.queries,
							m.redisClient,
							m.limiter,
							stopChan,
						)
					}
				}


				// Step 6: SCALE DOWN (Stop workers)
				if requiredWorkers < m.currentWorkers {

					workersToStop := m.currentWorkers - requiredWorkers

					log.Printf("Stopping %d workers", workersToStop)

					for i := 0; i < workersToStop; i++ {

						// Get stop channel of last worker
						stopChan := m.stopChans[m.currentWorkers]

						// Send stop signal to worker
						stopChan <- true

						// Remove worker from map
						delete(m.stopChans, m.currentWorkers)

						// Reduce worker count
						m.currentWorkers--
					}
				}

				// Step 7: Wait before next check
				// Manager checks queue every 5 seconds
				time.Sleep(5 * time.Second)
		}
	}
}

func (m *WorkerManager) Stop() {

	log.Println("Stopping all workers...")

	for id, stopChan := range m.stopChans {

		log.Println("Stopping worker:", id)

		stopChan <- true

		delete(m.stopChans, id)
	}
	m.currentWorkers = 0
}