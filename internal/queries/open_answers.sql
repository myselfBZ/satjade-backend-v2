-- name: CreateOpenAnswerKey :one
INSERT INTO open_answer_keys(question_id, match_pattern, model_answer) VALUES($1, $2, $3)
RETURNING *;
