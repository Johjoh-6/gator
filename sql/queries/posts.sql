-- name: CreatePost :one
INSERT INTO posts (
	id, created_at, updated_at,
	feed_id, title, url, description, published_at
) VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetPostForUser :many
SELECT posts.*, feeds.name AS feed_name FROM posts
JOIN feed_follows ON feed_follows.feed_id = posts.feed_id
JOIN feeds ON posts.feed_id = feeds.id
WHERE feed_follows.user_id = $1
ORDER BY posts.published_at DESC
LIMIT $2;
