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

-- What Decides Placeholder Number? POSITION inside the query.
