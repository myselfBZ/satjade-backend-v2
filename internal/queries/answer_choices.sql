-- name: CreateAnswerChoices :many
INSERT INTO answer_choices (question_id, label, body)
VALUES
    ($1, $2, $3),
    ($1, $4, $5),
    ($1, $6, $7),
    ($1, $8, $9)
RETURNING *;
