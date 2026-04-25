-- name: CreatePracticeWithModules :one
WITH new_practice AS (
    INSERT INTO practices (title)
    VALUES ($1)
    RETURNING id, title, created_at, updated_at
),
inserted_modules AS (
    INSERT INTO modules (practice_id, order_index, name)
    SELECT id, order_index, name
    FROM new_practice,
    (VALUES
        (1, 'Reading And Writing 1'),
        (2, 'Reading And Writing 2'),
        (3, 'Math 1'),
        (4, 'Math 2')
    ) AS t(order_index, name)
    RETURNING id

)
SELECT * FROM new_practice;


-- name: GetPracticePreviews :many
SELECT * FROM practices;


-- name: GetPracticeWithModules :many
SELECT
    p.id,
    p.title,
    p.created_at,
    p.updated_at,
    m.id AS module_id,
    m.name AS module_name,
    m.order_index
FROM practices p
JOIN modules m ON m.practice_id = p.id
WHERE p.id = $1
ORDER BY m.order_index;
