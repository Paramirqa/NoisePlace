-- name: AddArtistExternalID :one
INSERT INTO artist_external_ids (
    artist_id,
    source,
    external_id
)
VALUES (
    sqlc.arg(artist_id),
    sqlc.arg(source),
    sqlc.arg(external_id)
)
RETURNING *;


-- name: GetArtistByExternalID :one
SELECT a.*
FROM artists AS a
JOIN artist_external_ids AS e
    ON e.artist_id = a.id
WHERE e.source = sqlc.arg(source)
  AND e.external_id = sqlc.arg(external_id)
LIMIT 1;


-- name: AddReleaseExternalID :one
INSERT INTO release_external_ids (
    release_id,
    source,
    external_id
)
VALUES (
    sqlc.arg(release_id),
    sqlc.arg(source),
    sqlc.arg(external_id)
)
RETURNING *;


-- name: GetReleaseByExternalID :one
SELECT r.*
FROM releases AS r
JOIN release_external_ids AS e
    ON e.release_id = r.id
WHERE e.source = sqlc.arg(source)
  AND e.external_id = sqlc.arg(external_id)
LIMIT 1;