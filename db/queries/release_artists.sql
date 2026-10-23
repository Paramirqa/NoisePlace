-- name: AddReleaseArtist :exec
INSERT INTO release_artists (
    release_id,
    artist_id,
    role,
    position
)
VALUES (
    sqlc.arg(release_id),
    sqlc.arg(artist_id),
    sqlc.arg(role),
    sqlc.arg(position)
);


-- name: GetReleaseArtists :many
SELECT
    a.id,
    a.name,
    a.normalized_name,
    a.sort_name,
    ra.role,
    ra.position
FROM release_artists AS ra
JOIN artists AS a ON a.id = ra.artist_id
WHERE ra.release_id = sqlc.arg(release_id)
ORDER BY ra.position, a.name, a.id;


-- name: ListArtistReleases :many
SELECT
    r.*
FROM releases AS r
JOIN release_artists AS ra ON ra.release_id = r.id
WHERE ra.artist_id = sqlc.arg(artist_id)
ORDER BY r.release_date DESC NULLS LAST, r.title, r.id
LIMIT sqlc.arg(page_limit)::INTEGER
OFFSET sqlc.arg(page_offset)::INTEGER;


-- name: RemoveReleaseArtist :execrows
DELETE FROM release_artists
WHERE release_id = sqlc.arg(release_id)
  AND artist_id = sqlc.arg(artist_id);