-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users;

-- name: CreateUser :one
INSERT INTO users (name, description)
VALUES ($1, $2) RETURNING *;

-- name: CheckLoginExists :one
SELECT EXISTS (
    SELECT 1 FROM users_auth WHERE login = $1
);

-- name: GetUserAuth :one
SELECT * FROM users_auth WHERE login = $1 LIMIT 1;

-- name: CreateUserAuth :exec
INSERT INTO users_auth (user_id ,login, password_hash)
VALUES ($1, $2, $3);

-- name: UpdateUser :exec
UPDATE users
SET name = $2, description = $3
WHERE id = $1;

-- name: DeletUser :exec
DELETE FROM users WHERE id = $1;