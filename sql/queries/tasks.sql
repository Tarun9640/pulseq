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
