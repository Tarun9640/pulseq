package service

import (
	"context"

	"github.com/Tarun9640/pulseq/internal/db"
	"github.com/Tarun9640/pulseq/internal/repository"
)

// TaskService handles business logic related to tasks.
// Handler -> Service -> Repository -> DB
// Service layer lo manam rules, validations, future features add chestham.
type TaskService struct {
	repo *repository.TaskRepository // DB operations ni handle chestundi
}

// NewTaskService creates a new instance of TaskService.
// Dependency Injection pattern use chesthunam.
// Direct ga repository create cheyyatledu -> loose coupling.
func NewTaskService(r *repository.TaskRepository) *TaskService {
	return &TaskService{
		repo: r,
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

		// Always create tasks with "pending" status.
		// Worker later pick chesi process chestadu.
		Status: "pending",

		Payload: payload,
	}

	// Repository ni call chesi DB lo task create chesthunam.
	return s.repo.CreateTask(ctx, params)
}
