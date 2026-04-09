-- name: CreateQuestionAttempt :one
INSERT INTO question_attempts(response, elapsed_seconds ,question_id, user_id, is_correct) VALUES(
    $1,
    $2,
    $3,
    $4,
    $5
) RETURNING *;


-- name: GetQuestionAttemptsByUser :many
SELECT * FROM question_attempts WHERE user_id = $1;
