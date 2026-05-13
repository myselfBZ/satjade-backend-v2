-- name: CreateFriend :one
INSERT INTO friends(
    user1,
    user2
) VALUES($1, $2) RETURNING *;


-- name: DeleteFriend :exec
DELETE FROM friends WHERE id = $1; 

-- name: GetFriendsByUser :many
SELECT 
    users.id  AS friend_id,
    friends.id AS friendship_id,
    users.full_name, 
    friends.created_at 
FROM friends 
    JOIN users ON users.id = CASE 
        WHEN friends.user1 = $1 THEN friends.user2 
        ELSE friends.user1 
    END
WHERE friends.user1 = $1 OR friends.user2 = $1;


