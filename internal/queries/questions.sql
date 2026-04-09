-- name: CreateQuestion :one
INSERT INTO questions (
    type,
    paragraph,
    prompt,
    image_path,
    skill,
    domain,
    difficulty,
    explanation
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;


-- name: LinkQuestionToModule :exec
INSERT INTO module_questions (
    module_id,
    question_id,
    number
)  VALUES($1, $2, $3);

-- name: SetCorrectChoice :exec
UPDATE questions
SET correct_choice_id = $1
WHERE id = $2;


-- name: GetQuestionDistribution :many
SELECT 
    q.domain, 
    q.skill,
    q.difficulty,
    COUNT(q.id)                                          AS total_count,
    COUNT(qa.question_id) FILTER (WHERE qa.is_correct = true)  AS correct_count,
    COUNT(qa.question_id) FILTER (WHERE qa.is_correct = false) AS incorrect_count,
    COUNT(q.id) FILTER (WHERE qa.question_id IS NULL)          AS unattempted_count
FROM questions q
LEFT JOIN question_attempts qa ON qa.question_id = q.id AND qa.user_id = $1
GROUP BY q.domain, q.skill, q.difficulty;

-- name: GetQuestionsByModule :many
SELECT
    q.id,
    q.type,
    q.paragraph,
    q.prompt,
    q.image_path,
    q.correct_choice_id,
    q.difficulty,
    q.domain,
    q.skill,
    q.explanation,
    mq.number,

    (SELECT json_agg(ch ORDER BY ch.label) 
        FROM (
            SELECT id, label, body 
            FROM answer_choices 
            WHERE question_id = q.id
        ) ch
    ) 
    AS choices,

    (SELECT json_build_object(
        'id', oak.id,
        'model_answer', oak.model_answer,
        'match_pattern', oak.match_pattern
     ) 
     FROM open_answer_keys oak 
     WHERE oak.question_id = q.id
    ) AS open_key
FROM questions q
JOIN module_questions mq ON q.id = mq.question_id
WHERE mq.module_id = $1
ORDER BY mq.number;




-- name: FilterQuestions :many
SELECT q.id
FROM questions q
LEFT JOIN LATERAL (
    SELECT is_correct
    FROM question_attempts
    WHERE question_id = q.id
      AND user_id = @user_id::uuid
    ORDER BY created_at DESC
    LIMIT 1
) qa ON true
WHERE 
    (difficulty = ANY(@difficulty_levels::difficulty[]))
    AND (
        @domains::text[] @> ARRAY['all']::text[]
        OR domain = ANY(@domains::text[])
    )
    AND (
        @skills::text[] @> ARRAY['all']::text[] 
        OR skill = ANY(@skills::text[])
    )
    AND (
        @attempt_status::text = 'all'
        OR (@attempt_status::text = 'correct'     AND qa.is_correct = true)
        OR (@attempt_status::text = 'incorrect'   AND qa.is_correct = false)
        OR (@attempt_status::text = 'unattempted' AND qa.is_correct IS NULL)
    );


-- name: GetQuestionByID :one
SELECT
    q.id,
    q.type,
    q.paragraph,
    q.prompt,
    q.image_path,
    q.skill,
    q.domain,
    q.difficulty,
    q.explanation,
    q.correct_choice_id,
    json_agg(
        json_build_object(
            'id',    ac.id,
            'label', ac.label,
            'body',  ac.body
        )
    ) FILTER (WHERE ac.id IS NOT NULL)
    AS choices,
    oak.id            AS answer_key_id,
    oak.model_answer  AS answer_key_model_answer,
    oak.match_pattern AS answer_key_match_pattern
FROM questions q
LEFT JOIN answer_choices ac  ON ac.question_id = q.id
LEFT JOIN open_answer_keys oak ON oak.question_id = q.id
WHERE q.id = $1
GROUP BY q.id, oak.id;
