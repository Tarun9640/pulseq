package worker

import (
	"context"
	"errors"
	"log"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/Tarun9640/pulseq/internal/db"
	"github.com/Tarun9640/pulseq/internal/queue"
	"github.com/google/uuid"
	redisLib "github.com/redis/go-redis/v9"
)

// flow
// 1️⃣ BRPOPLPUSH
// 2️⃣ Fetch DB task
// 3️⃣ If completed → skip
// 4️⃣ If next_retry_at > now → requeue + skip
// 5️⃣ Process
// 6️⃣ If fail → schedule retry or DLQ
// 7️⃣ If success → ACK + mark completed


func StartWorker(workerID int, queries *db.Queries, redisClient *redisLib.Client) {

	ctx := context.Background()

	log.Printf("Worker %d started...\n", workerID)

	for {
		// Move job safely from main -> processing
		//BRPOP = Blocking Right Pop Worker waits…Until job arrives.
		//0 = wait FOREVER if BRPOP ctx, 5, queue wait for 5 sec if no job return error
		taskID, err := redisClient.BRPopLPush(ctx, queue.TaskQueueName, queue.ProcessingQueueName, 0).Result() 
		
		if err != nil {
			log.Println("Worker error:", err)
			continue
		}

		log.Printf("Worker %d picked task %s\n", workerID, taskID)

		// Parse UUID
		parsedID, err := uuid.Parse(taskID)
		if err != nil {
			log.Println("invalid uuid:", err)

			// remove corrupted job
			redisClient.LRem(ctx, queue.ProcessingQueueName, 1, taskID)
			continue
		}


		// Fetch task from DB
		task, err := queries.GetTask(ctx, parsedID)
		if err != nil {
			log.Println("DB fetch error:", err)
			continue
		}

		log.Printf("Worker %d processing task type: %s\n", workerID, task.Type)

		// idempotency -- is status is already completed worker dont execute
		if task.Status == "completed" {
			log.Printf("Task %s already completed, skipping\n", taskID)
			redisClient.LRem(ctx, queue.ProcessingQueueName, 1, taskID)
			continue
		}

		// simulate work 
		err = processTask(task)

		if err != nil {
			log.Printf("Worker %d failed task %s: %v\n", workerID, taskID, err)

			//retry logic
			// for exponential backoff:
			// Calculate delay
			// Store next_retry_at in DB
			// Requeue
			// Worker before processing check cheyali → time ayinda leda

			baseDelay := 2 * time.Second

			delayTime := baseDelay * (time.Duration(math.Pow(2, float64(task.RetryCount))))

			nextRetry := time.Now().Add(delayTime)

			//step 2:  Store next_retry_at in DB
			rows, err := queries.ScheduleRetry(ctx, db.ScheduleRetryParams{
				ID : task.ID,
				NextRetryAt: &nextRetry,
			})

			if err != nil {
				log.Println("next_retry_at update failed in db:", err)
				redisClient.LRem(ctx, queue.ProcessingQueueName, 1, taskID)
				continue
			}

			if rows == 0 {
				// Retry exhausted → DLQ
				log.Printf("Task %s moved to DLQ\n", taskID)

				redisClient.LPush(ctx, queue.DeadLetterQueue, taskID)
				queries.MoveToFailed(ctx, task.ID)
			} else {
				// Retry allowed → Delay Queue
				log.Printf("Task %s scheduled for retry\n", taskID)

				redisClient.ZAdd(ctx, queue.DelayQueue, redisLib.Z{
					Score: float64(nextRetry.Unix()),
					Member: taskID,
				})

			}
			// Always ACK processing queue
			redisClient.LRem(ctx, queue.ProcessingQueueName, 1, taskID)
			continue
		}
		
		//--------------success flow -----------------

		// we update status as completed
		err = queries.UpdateTaskStatus(ctx, db.UpdateTaskStatusParams{
			ID: task.ID,
			Status: "completed",
		})

		if err != nil {
			log.Println("status update failed:", err)
			continue
		}
		// Because Redis list doesn’t auto-delete processed items. YOU must acknowledge completion.
		// This is called: ACK Pattern (Acknowledgement)
		//manually remove completed task from queue
		_,err = redisClient.LRem(ctx, queue.ProcessingQueueName, 1, taskID).Result() 
		
		if err != nil  {
			log.Println("Failed to ACK task:", err)
			continue
		}

		log.Printf("Worker %d completed task %s\n", workerID, taskID)
	}
}

func processTask(task db.Task) error {

	//simulate random failure
	if rand.Intn(5) == 0 {
		return errors.New("random failure occurred")
	}

	time.Sleep(2 * time.Second)

	return nil

	//return errors.New("forced failure")
}


func Scheduler(redisClient *redisLib.Client) {
	log.Println("Scheduler started...")
	ctx := context.Background()
	for {
		now := time.Now().Unix()

		tasks, err := redisClient.ZRangeByScore(ctx, queue.DelayQueue, &redisLib.ZRangeBy{
			Min: "-inf",
			Max: strconv.FormatInt(now, 10),
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

			log.Printf("Scheduler moved task %s to main queue\n", taskID)
		}
		time.Sleep(1 * time.Second)
	}
}

