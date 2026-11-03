-- name: UpsertArtistRating :one
INSERT INTO user_artist_ratings (
    user_id,
    artist_id,
    rating
)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, artist_id)
DO UPDATE SET
    rating = EXCLUDED.rating,
    updated_at = NOW()
RETURNING *;


-- name: DeleteArtistRating :execrows
DELETE FROM user_artist_ratings
WHERE user_id = $1
  AND artist_id = $2;


-- name: GetArtistRating :one
SELECT *
FROM user_artist_ratings
WHERE user_id = $1
  AND artist_id = $2;


-- name: GetArtistRatingStats :one
SELECT
    COALESCE(ROUND(AVG(rating)::numeric, 2), 0)::numeric(4, 2) AS average_rating,
    COUNT(*)::bigint AS ratings_count
FROM user_artist_ratings
WHERE artist_id = $1;


-- name: UpsertReleaseRating :one
INSERT INTO user_release_ratings (
    user_id,
    release_id,
    rating
)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, release_id)
DO UPDATE SET
    rating = EXCLUDED.rating,
    updated_at = NOW()
RETURNING *;


-- name: DeleteReleaseRating :execrows
DELETE FROM user_release_ratings
WHERE user_id = $1
  AND release_id = $2;


-- name: GetReleaseRating :one
SELECT *
FROM user_release_ratings
WHERE user_id = $1
  AND release_id = $2;


-- name: GetReleaseRatingStats :one
SELECT
    COALESCE(ROUND(AVG(rating)::numeric, 2), 0)::numeric(4, 2) AS average_rating,
    COUNT(*)::bigint AS ratings_count
FROM user_release_ratings
WHERE release_id = $1;