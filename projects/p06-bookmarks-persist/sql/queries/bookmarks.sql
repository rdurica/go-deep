-- name: InsertBookmark :exec
INSERT INTO bookmarks (id, url, title, tags, created_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetBookmark :one
SELECT id, url, title, tags, created_at
FROM bookmarks
WHERE id = $1;

-- name: DeleteBookmark :execrows
DELETE FROM bookmarks
WHERE id = $1;

-- name: SearchBookmarks :many
SELECT id, url, title, tags, created_at
FROM bookmarks
WHERE
    (sqlc.arg(query_text)::text = '' OR title ILIKE '%' || sqlc.arg(query_text) || '%' OR url ILIKE '%' || sqlc.arg(query_text) || '%')
    AND (sqlc.arg(tag)::text = '' OR sqlc.arg(tag) = ANY (tags))
ORDER BY created_at DESC, id ASC;
