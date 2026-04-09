-- name: PublishPractice :one
INSERT INTO published_practices(practice_id, published_by, data) VALUES(
    $1, $2, $3
) RETURNING id;

-- name: GetPublishedPractice :one
SELECT * FROM published_practices WHERE id = $1;
