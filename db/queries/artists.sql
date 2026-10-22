-- name: CreateArtist :one
INSERT INTO artists (
    name,
    normalized_name,
    sort_name,
    description,
    image_url
)
VALUES (
    sqlc.arg(name),
    sqlc.arg(normalized_name),
    sqlc.narg(sort_name),
    sqlc.narg(description),
    sqlc.narg(image_url)
)
RETURNING *;


-- name: GetArtist :one
SELECT *
FROM artists
WHERE id = sqlc.arg(id)
LIMIT 1;


-- name: SearchArtists :many
SELECT *
FROM artists
WHERE normalized_name LIKE '%' || sqlc.arg(query)::TEXT || '%'
ORDER BY normalized_name, id
LIMIT sqlc.arg(page_limit)::INTEGER
OFFSET sqlc.arg(page_offset)::INTEGER;


-- name: UpdateArtist :one
UPDATE artists
SET
    name = sqlc.arg(name),
    normalized_name = sqlc.arg(normalized_name),
    sort_name = sqlc.narg(sort_name),
    description = sqlc.narg(description),
    image_url = sqlc.narg(image_url),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;