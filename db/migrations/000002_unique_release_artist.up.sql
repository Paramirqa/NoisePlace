-- Keep one relationship per release and artist.
-- When duplicates exist, prefer primary, then the lowest position.

WITH ranked AS (
    SELECT
        release_id,
        artist_id,
        role,
        ROW_NUMBER() OVER (
            PARTITION BY release_id, artist_id
            ORDER BY
                CASE role
                    WHEN 'primary' THEN 0
                    WHEN 'featured' THEN 1
                    WHEN 'producer' THEN 2
                    ELSE 3
                END,
                position,
                role
        ) AS row_num
    FROM release_artists
)
DELETE FROM release_artists ra
USING ranked r
WHERE ra.release_id = r.release_id
  AND ra.artist_id = r.artist_id
  AND ra.role = r.role
  AND r.row_num > 1;

ALTER TABLE release_artists
    DROP CONSTRAINT release_artists_pkey;

ALTER TABLE release_artists
    ADD CONSTRAINT release_artists_pkey
    PRIMARY KEY (release_id, artist_id);