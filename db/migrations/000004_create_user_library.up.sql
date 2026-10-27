CREATE TABLE user_library_artists (
    user_id BIGINT NOT NULL,
    artist_id BIGINT NOT NULL,
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (user_id, artist_id),

    CONSTRAINT user_library_artists_user_fk
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE CASCADE,

    CONSTRAINT user_library_artists_artist_fk
        FOREIGN KEY (artist_id)
        REFERENCES artists (id)
        ON DELETE CASCADE
);

CREATE INDEX idx_user_library_artists_artist_id
    ON user_library_artists (artist_id);


CREATE TABLE user_library_releases (
    user_id BIGINT NOT NULL,
    release_id BIGINT NOT NULL,
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (user_id, release_id),

    CONSTRAINT user_library_releases_user_fk
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE CASCADE,

    CONSTRAINT user_library_releases_release_fk
        FOREIGN KEY (release_id)
        REFERENCES releases (id)
        ON DELETE CASCADE
);

CREATE INDEX idx_user_library_releases_release_id
    ON user_library_releases (release_id);