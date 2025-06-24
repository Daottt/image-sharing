-- name: GetSession :one
SELECT * FROM sessions WHERE id = $1;

-- name: CreateSession :one
INSERT INTO sessions (id, user_login, access_token, refresh_token, is_revoked, expires_at)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: RevokeSession :exec
UPDATE sessions
SET is_revoked = TRUE
WHERE access_token = $1;

-- name: RevokeSessionsByLogin :exec
UPDATE sessions
SET is_revoked = TRUE
WHERE user_login = $1;

-- name: RenewSession :exec
UPDATE sessions
SET access_token = $2
WHERE id = $1;

-- name: DeletSession :exec
DELETE FROM sessions WHERE id = $1;