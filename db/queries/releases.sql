-- name: CreateRelease :one
INSERT INTO releases (
title,
release_type,
release_date,
description,
cover_url
)
VALUES (
sqlc.arg(title),
sqlc.arg(release_type),
sqlc.narg(release_date),
sqlc.narg(description),
sqlc.narg(cover_url)
)
RETURNING *;


-- name: GetRelease :one
SELECT *
FROM releases
WHERE id = sqlc.arg(id)
LIMIT 1;

-- name: SearchReleases :many
SELECT *
FROM releases
WHERE title ILIKE '%' || sqlc.arg(query)::TEXT || '%'
ORDER BY release_date DESC NULLS LAST, title, id
LIMIT sqlc.arg(page_limit)::INTEGER
OFFSET sqlc.arg(page_offset)::INTEGER;

-- name: UpdateRelease :one
UPDATE releases
SET
title = sqlc.arg(title),
release_type = sqlc.arg(release_type),
release_date = sqlc.narg(release_date),
description = sqlc.narg(description),
cover_url = sqlc.narg(cover_url),
updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;