package worker

import (
	"context"
	"log"
	"log/slog"
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
	logger *slog.Logger
}

func NewWorkerManager(client *redis.Client, limiter *ratelimiter.TokenBucketLimiter, queries *db.Queries, minWorkers int, maxWorkers int, logger *slog.Logger) *WorkerManager {
	return &WorkerManager{
		redisClient: client,
		limiter: limiter,
		queries: queries,
		minWorkers: minWorkers,
		maxWorkers: maxWorkers,
		currentWorkers: 0,
		stopChans: make(map[int]chan bool),
		logger: logger,
	}
}

func (m *WorkerManager) Start(ctx context.Context) {

	m.logger.Info("Worker Manager started...")

	for {

		select {
			case <-ctx.Done():
				m.logger.Info("Worker manager stopping")
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

				m.logger.Info("queue metrics",
					"queue_size", queueSize,
					"workers", m.currentWorkers,
				)

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

				m.logger.Info("worker health",
					"active_workers", m.currentWorkers,
					"min_workers", m.minWorkers,
					"max_workers", m.maxWorkers,
				)


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
							m.logger,
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

	m.logger.Info("Stopping all workers...")

	for id, stopChan := range m.stopChans {

		m.logger.Info("Stopping worker:", "worker_id", id)

		stopChan <- true

		delete(m.stopChans, id)
	}
	m.currentWorkers = 0
}

func (m *WorkerManager) GetWorkerCount() int {
	return m.currentWorkers
}

func (m *WorkerManager) GetWorkerStats(ctx context.Context) (map[string]interface{}, error) {

	queueSize, err := m.redisClient.LLen(ctx, queue.TaskQueueName).Result()

	if err !=nil {
		return nil, err
	}

	stats := map[string]interface{} {
		"active_workers" : m.currentWorkers,
		"min_workers" : m.minWorkers,
		"max_workers" : m.maxWorkers,
		"queue_depth" : queueSize,
	}

	return stats, nil
}