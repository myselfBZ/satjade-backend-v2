-- name: GetModulesByPracticeID :many
SELECT id, name, order_index FROM modules WHERE practice_id = $1 ORDER BY order_index;
