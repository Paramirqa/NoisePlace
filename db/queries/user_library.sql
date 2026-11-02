-- name: AddArtistToLibrary :exec
INSERT INTO user_library_artists (user_id, artist_id)
VALUES (sqlc.arg(user_id), sqlc.arg(artist_id))
ON CONFLICT (user_id, artist_id) DO NOTHING;


-- name: RemoveArtistFromLibrary :execrows
DELETE FROM user_library_artists
WHERE user_id = sqlc.arg(user_id)
  AND artist_id = sqlc.arg(artist_id);


-- name: AddReleaseToLibrary :exec
INSERT INTO user_library_releases (user_id, release_id)
VALUES (sqlc.arg(user_id), sqlc.arg(release_id))
ON CONFLICT (user_id, release_id) DO NOTHING;


-- name: RemoveReleaseFromLibrary :execrows
DELETE FROM user_library_releases
WHERE user_id = sqlc.arg(user_id)
  AND release_id = sqlc.arg(release_id);

-- name: ListLibraryArtists :many
SELECT
    a.id,
    a.name,
    a.sort_name,
    ula.added_at
FROM user_library_artists AS ula
JOIN artists AS a ON a.id = ula.artist_id
WHERE ula.user_id = sqlc.arg(user_id)
ORDER BY ula.added_at DESC, a.name ASC;

-- name: ListLibraryReleases :many
SELECT
    r.id,
    r.title,
    r.release_type,
    r.release_date,
    r.cover_url,
    ulr.added_at
FROM user_library_releases AS ulr
JOIN releases AS r ON r.id = ulr.release_id
WHERE ulr.user_id = sqlc.arg(user_id)
ORDER BY ulr.added_at DESC, r.title ASC;