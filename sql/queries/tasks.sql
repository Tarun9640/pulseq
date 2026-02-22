-- name: CreateTask :one
INSERT INTO tasks (
    type,
    status,
    payload
) VALUES (
    $1, $2, $3
)
RETURNING *;


-- name: GetTask :one
SELECT * FROM tasks
WHERE id = $1;

-- name: UpdateTaskStatus :exec
UPDATE tasks
SET status = $2, updated_at = NOW()
WHERE id = $1;

-- What Decides Placeholder Number(like $1,$2)? POSITION inside the query.

-- name: IncrementRetry :exec
UPDATE tasks
SET retry_count = retry_count + 1, updated_at = NOW()
WHERE id = $1;

-- name: MoveToFailed :exec
UPDATE tasks
SET status = 'failed',
    updated_at = NOW()
WHERE id = $1;


-- name: ScheduleRetry :execrows
UPDATE tasks
SET retry_count = retry_count + 1,
    next_retry_at = $2,
    updated_at = NOW()
WHERE id = $1 AND retry_count < max_retries;
