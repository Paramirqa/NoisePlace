ALTER TABLE release_artists
    DROP CONSTRAINT release_artists_pkey;

ALTER TABLE release_artists
    ADD CONSTRAINT release_artists_pkey
    PRIMARY KEY (release_id, artist_id, role);