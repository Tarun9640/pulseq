package worker

import (
	"context"
	"log"
	"time"

	"github.com/Tarun9640/pulseq/internal/db"
	"github.com/Tarun9640/pulseq/internal/queue"
	"github.com/google/uuid"
	redisLib "github.com/redis/go-redis/v9"
)

func StartWorker(id int, queries *db.Queries, redisClient *redisLib.Client) {

	ctx := context.Background()

	for {
		// BLOCK until job arrives
		result, err := redisClient.BRPopLPush(ctx, queue.TaskQueueName, queue.ProcessingQueueName, 0).Result() //BRPOP = Blocking Right Pop Worker waits…Until job arrives.
		//0 = wait FOREVER if BRPOP ctx, 5, queue wait for 5 sec if no job return error
		if err != nil {
			log.Println("Worker error:", err)
			continue
		}

		log.Println(result)

		//taskID := result[1]

		log.Printf("Worker %d picked task %s\n", id, result)

		parsedID, err := uuid.Parse(result)
		if err != nil {
			log.Println("invalid uuid:", err)
			continue
		}


		// Fetch task from DB
		task, err := queries.GetTask(ctx, parsedID)
		if err != nil {
			log.Println("DB fetch error:", err)
			continue
		}

		log.Printf("Worker %d processing type: %s\n", id, task.Type)


		// simulate work -- at present we sleep for 2 sec
		time.Sleep(2 * time.Second)
		
		// later we update status
		err = queries.UpdateTaskStatus(ctx, db.UpdateTaskStatusParams{
			ID: task.ID,
			Status: "completed",
		})

		if err != nil {
			log.Println("status update failed:", err)
			continue
		}
		//Because Redis list doesn’t auto-delete processed items.YOU must acknowledge completion.
		// This is called: ACK Pattern (Acknowledgement)

		_,err = redisClient.LRem(ctx, queue.ProcessingQueueName, 1, result).Result()
		
		if err != nil {
			log.Println("Failed to remove from processing queue:", err)
		}

		log.Printf("Worker %d completed task\n", id)

	}
}
