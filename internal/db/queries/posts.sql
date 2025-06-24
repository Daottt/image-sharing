-- name: GetPost :one
SELECT posts.*, users.name as user_name FROM posts 
INNER JOIN users ON posts.user_id = users.id
WHERE posts.id = $1;

-- name: ListPosts :many
SELECT posts.*, users.name as user_name FROM posts 
INNER JOIN users ON posts.user_id = users.id
ORDER BY posts.id LIMIT $1 OFFSET $2;

-- name: CountPosts :one
SELECT count(*) FROM posts;

-- name: CreatePost :one
INSERT INTO posts (user_id, image_path)
VALUES ($1, $2) RETURNING *;

-- name: GetPostUserID :one
SELECT user_id FROM posts WHERE posts.id = $1;

-- name: DeletPost :exec
DELETE FROM posts WHERE id = $1;