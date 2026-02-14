package repository

import (
	"context"

	"github.com/Tarun9640/pulseq/internal/db"
)

// TaskRepository acts as the data access layer for tasks.
//
// Responsibility:
// This layer ONLY talks to the database.
// No business logic should be written here.
//
// Why this layer exists?
// → To decouple database code from service logic.
// → If DB technology changes tomorrow (Postgres → DynamoDB),
//   only this layer needs modification.
//
// Architecture Flow:
// Handler → Service → Repository → DB
type TaskRepository struct {
	queries *db.Queries // sqlc generated Queries struct (type-safe DB methods)
}


// NewTaskRepository creates a new repository instance.
//
// We inject *db.Queries instead of creating it inside,
// because dependency injection makes the code:
//
// ✅ testable
// ✅ modular
// ✅ easy to replace
//
// This is a common pattern in production Go services.
func NewTaskRepository(q *db.Queries) *TaskRepository {
	return &TaskRepository{
		queries: q,
	}
}


// CreateTask inserts a new task into the database.
//
// Why repository should handle this?
// → Service layer should not worry about SQL.
// → Repository abstracts DB interaction.
//
// Parameters:
// ctx     → Controls timeout/cancellation (VERY important in production)
// params  → sqlc-generated struct containing task fields
//
// Returns:
// db.Task → The inserted task (with auto-generated UUID, timestamps)
// error   → Any DB failure
//
// NOTE:
// This function is intentionally thin.
// Repository should stay lightweight and predictable.
func (r *TaskRepository) CreateTask(ctx context.Context, params db.CreateTaskParams,) (db.Task, error) {
	return r.queries.CreateTask(ctx, params)
}
