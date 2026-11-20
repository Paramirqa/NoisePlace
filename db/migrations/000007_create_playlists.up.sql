CREATE TABLE playlists (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT playlists_user_fk
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE CASCADE,

    CONSTRAINT playlists_title_length_check
        CHECK (char_length(BTRIM(title)) BETWEEN 1 AND 100),

    CONSTRAINT playlists_description_length_check
        CHECK (
            description IS NULL
            OR char_length(description) <= 5000
        )
);

CREATE INDEX idx_playlists_user_id
    ON playlists (user_id);


CREATE TABLE playlist_releases (
    playlist_id BIGINT NOT NULL,
    release_id BIGINT NOT NULL,
    position INTEGER NOT NULL,
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (playlist_id, release_id),

    CONSTRAINT playlist_releases_playlist_fk
        FOREIGN KEY (playlist_id)
        REFERENCES playlists (id)
        ON DELETE CASCADE,

    CONSTRAINT playlist_releases_release_fk
        FOREIGN KEY (release_id)
        REFERENCES releases (id)
        ON DELETE CASCADE,

    CONSTRAINT playlist_releases_position_check
        CHECK (position >= 0),

    CONSTRAINT playlist_releases_position_unique
        UNIQUE (playlist_id, position)
        DEFERRABLE INITIALLY DEFERRED
);

CREATE INDEX idx_playlist_releases_release_id
    ON playlist_releases (release_id);