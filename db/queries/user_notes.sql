
-- Artist notes

-- name: CreateArtistNote :exec
INSERT INTO user_artist_notes (user_id, artist_id, content)
VALUES (
    sqlc.arg(user_id),
    sqlc.arg(artist_id),
    sqlc.arg(content)
);

-- name: GetArtistNote :one
SELECT
    user_id,
    artist_id,
    content,
    created_at,
    updated_at
FROM user_artist_notes
WHERE user_id = sqlc.arg(user_id)
  AND artist_id = sqlc.arg(artist_id);

-- name: UpdateArtistNote :execrows
UPDATE user_artist_notes
SET
    content = sqlc.arg(content),
    updated_at = NOW()
WHERE user_id = sqlc.arg(user_id)
  AND artist_id = sqlc.arg(artist_id);

-- name: DeleteArtistNote :execrows
DELETE FROM user_artist_notes
WHERE user_id = sqlc.arg(user_id)
  AND artist_id = sqlc.arg(artist_id);


-- Release notes

-- name: CreateReleaseNote :exec
INSERT INTO user_release_notes (user_id, release_id, content)
VALUES (
    sqlc.arg(user_id),
    sqlc.arg(release_id),
    sqlc.arg(content)
);

-- name: GetReleaseNote :one
SELECT
    user_id,
    release_id,
    content,
    created_at,
    updated_at
FROM user_release_notes
WHERE user_id = sqlc.arg(user_id)
  AND release_id = sqlc.arg(release_id);

-- name: UpdateReleaseNote :execrows
UPDATE user_release_notes
SET
    content = sqlc.arg(content),
    updated_at = NOW()
WHERE user_id = sqlc.arg(user_id)
  AND release_id = sqlc.arg(release_id);

-- name: DeleteReleaseNote :execrows
DELETE FROM user_release_notes
WHERE user_id = sqlc.arg(user_id)
  AND release_id = sqlc.arg(release_id);
