-- name: GetUserById :one
SELECT  * FROM users WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users(email, role, full_name, password_hash) VALUES($1, $2, $3, $4) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUsers :many
SELECT * FROM users;

-- name: DeleteUser :one
DELETE FROM users WHERE id = $1 RETURNING *;
