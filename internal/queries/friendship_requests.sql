-- name: CreateFriendshipRequest :one
INSERT INTO friendship_requests (
    to_id,
    from_id,
    message
) VALUES ($1, $2, $3)
RETURNING 
    id, 
    to_id, 
    from_id, 
    message, 
    created_at,
    (SELECT full_name FROM users WHERE users.id = friendship_requests.from_id) AS sender_full_name;

-- name: DeleteFriendshipRequest :one
DELETE FROM friendship_requests WHERE id = $1 RETURNING *;

-- name: GetFriendshipRequestsByUser :many
SELECT 
    users.full_name, 
    friendship_requests.*
FROM friendship_requests JOIN users ON friendship_requests.from_id = users.id WHERE to_id = $1;
