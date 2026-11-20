-- name: CreatePlaylist :one
INSERT INTO playlists (
    user_id,
    title,
    description
)
VALUES (
    $1,
    $2,
    $3
)
RETURNING
    id,
    user_id,
    title,
    description,
    is_public,
    created_at,
    updated_at;

-- name: ListUserPlaylists :many
SELECT
    id,
    user_id,
    title,
    description,
    is_public,
    created_at,
    updated_at
FROM playlists
WHERE user_id = $1
ORDER BY created_at DESC, id DESC;

-- name: GetUserPlaylist :one
SELECT
    id,
    user_id,
    title,
    description,
    is_public,
    created_at,
    updated_at
FROM playlists
WHERE id = $1
  AND user_id = $2;

-- name: UpdatePlaylist :one
UPDATE playlists
SET
    title = COALESCE(sqlc.narg(title)::text, title),
    description = CASE
        WHEN sqlc.arg(update_description)::boolean
            THEN sqlc.narg(description)::text
        ELSE description
    END,
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id)
RETURNING
    id,
    user_id,
    title,
    description,
    is_public,
    created_at,
    updated_at;

-- name: DeletePlaylist :execrows
DELETE FROM playlists
WHERE id = $1
  AND user_id = $2; 

-- name: GetPlaylistForUpdate :one
SELECT id
FROM playlists
WHERE id = $1
  AND user_id = $2
FOR UPDATE;

-- name: AddReleaseToPlaylist :one
INSERT INTO playlist_releases (
    playlist_id,
    release_id,
    position
)
SELECT
    sqlc.arg(playlist_id),
    sqlc.arg(release_id),
    COALESCE(MAX(position), -1) + 1
FROM playlist_releases
WHERE playlist_id = sqlc.arg(playlist_id)
RETURNING
    playlist_id,
    release_id,
    position,
    added_at;

-- name: RemoveReleaseFromPlaylist :execrows
DELETE FROM playlist_releases
WHERE playlist_id = $1
  AND release_id = $2
  AND EXISTS (
      SELECT 1
      FROM playlists
      WHERE playlists.id = playlist_releases.playlist_id
        AND playlists.user_id = $3
  );

-- name: ListPlaylistReleases :many
SELECT
    r.id,
    r.title,
    r.release_type,
    r.release_date,
    pr.position,
    pr.added_at
FROM playlist_releases pr
JOIN releases r
    ON r.id = pr.release_id
JOIN playlists p
    ON p.id = pr.playlist_id
WHERE p.id = $1
  AND p.user_id = $2
ORDER BY pr.position ASC;

-- name: ListPlaylistReleaseIDs :many
SELECT release_id
FROM playlist_releases
WHERE playlist_id = $1
ORDER BY position;

-- name: UpdatePlaylistReleasePosition :execrows
UPDATE playlist_releases
SET position = $3
WHERE playlist_id = $1
  AND release_id = $2;

