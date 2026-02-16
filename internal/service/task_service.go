package service

import (
	"context"
	"log"

	"github.com/Tarun9640/pulseq/internal/db"
	"github.com/Tarun9640/pulseq/internal/queue"
	"github.com/Tarun9640/pulseq/internal/repository"
)

// TaskService handles business logic related to tasks.
// Handler -> Service -> Repository -> DB
// Service layer lo manam rules, validations, future features add chestham.
type TaskService struct {
	repo *repository.TaskRepository // DB operations ni handle chestundi
	queue *queue.RedisQueue
}

// NewTaskService creates a new instance of TaskService.
// Dependency Injection pattern use chesthunam.
// Direct ga repository create cheyyatledu -> loose coupling.
func NewTaskService(r *repository.TaskRepository, q *queue.RedisQueue) *TaskService {
	return &TaskService{
		repo: r,
		queue: q,
	}
}

// CreateTask creates a new task in the system.
//
// Flow:
// Handler -> calls Service
// Service -> prepares params + applies business rules
// Repository -> executes DB query
//
// ctx:
// request lifecycle control kosam use avuthundi.
// Example: request cancel ayithe DB query kuda stop avvali.
func (s *TaskService) CreateTask(ctx context.Context, taskType string, payload []byte,) (db.Task, error) {

	// sqlc generated struct.
	// DB insert ki required fields prepare chesthunam.
	params := db.CreateTaskParams{
		Type: taskType,
		Status: "pending", // Always create tasks with "pending" status.Worker later pick chesi process chestadu.
		Payload: payload,
	}

	// ✅ Step 1 — Insert into DB
	task, err := s.repo.CreateTask(ctx, params)
	if err != nil {
		return task, err
	}

	// ✅ Step 2 — Push to Redis queue
	log.Println("Pushing task to Redis:", task.ID.String())

	err = s.queue.Enqueue(ctx, task.ID.String())
	if err != nil {
		return task, err
	}
	log.Println("Task pushed successfully")


	// ✅ Step 3 — Return task
	return task, nil
}
