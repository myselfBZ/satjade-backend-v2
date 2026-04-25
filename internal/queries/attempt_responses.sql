-- name: CreateAttemptResponses :exec
WITH data AS (
    SELECT
        unnest($1::UUID[])    AS attempt_id,
        unnest($2::UUID[])    AS question_id,
        unnest($3::UUID[])    AS selected_choice_id,
        NULLIF(unnest($4::TEXT[]), '')    AS open_response,
        unnest($5::BOOLEAN[]) AS is_correct
)
INSERT INTO attempt_responses (
    attempt_id, question_id, selected_choice_id, open_response, is_correct
)
SELECT attempt_id, question_id, selected_choice_id, open_response, is_correct
FROM data;


-- name: GetAttemptResponsesByAttempt :many
SELECT * FROM attempt_responses WHERE attempt_id = $1;
