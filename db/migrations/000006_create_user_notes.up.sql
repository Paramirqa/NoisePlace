
CREATE TABLE user_artist_notes (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    artist_id BIGINT NOT NULL REFERENCES artists(id) ON DELETE CASCADE,
    content TEXT NOT NULL
        CHECK (char_length(BTRIM(content)) BETWEEN 1 AND 5000),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (user_id, artist_id)
);

CREATE INDEX idx_user_artist_notes_artist_id
    ON user_artist_notes (artist_id);


CREATE TABLE user_release_notes (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    release_id BIGINT NOT NULL REFERENCES releases(id) ON DELETE CASCADE,
    content TEXT NOT NULL
        CHECK (char_length(BTRIM(content)) BETWEEN 1 AND 5000),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (user_id, release_id)
);

CREATE INDEX idx_user_release_notes_release_id
    ON user_release_notes (release_id);
