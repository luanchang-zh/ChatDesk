-- 所有会话 SQL 都带 user_id。找不到行时调用方当成「不存在」，不区分是否属于别人。

-- name: EnsureUser :exec
INSERT INTO users (id) VALUES ($1) ON CONFLICT (id) DO NOTHING;

-- name: CreateConversation :one
INSERT INTO conversations (id, user_id, title) VALUES ($1, $2, $3)
RETURNING *;

-- name: GetConversation :one
SELECT * FROM conversations WHERE id = $1 AND user_id = $2;

-- name: ListConversations :many
SELECT * FROM conversations WHERE user_id = $1
ORDER BY updated_at DESC, id DESC;

-- name: RenameConversation :one
UPDATE conversations SET title = $3, updated_at = clock_timestamp()
WHERE id = $1 AND user_id = $2 RETURNING *;

-- name: DeleteConversation :execrows
DELETE FROM conversations WHERE id = $1 AND user_id = $2;
