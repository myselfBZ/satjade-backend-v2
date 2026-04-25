-- name: CreatePracticeAttempt :one
INSERT INTO practice_attempts(
    user_id, 
    published_practice_id, 
    started_at, 
    submitted_at, 
    rw_score, 
    math_score) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;


-- name: GetPracticeAttemptsByUser :many
SELECT * FROM practice_attempts WHERE user_id = $1 ORDER BY submitted_at DESC;


-- name: GetById :one
SELECT * FROM practice_attempts WHERE id = $1;
