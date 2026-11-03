CREATE TABLE user_artist_ratings (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    artist_id BIGINT NOT NULL REFERENCES artists(id) ON DELETE CASCADE,
    rating SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 10),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (user_id, artist_id)
);

CREATE INDEX idx_user_artist_ratings_artist_id
    ON user_artist_ratings (artist_id);


CREATE TABLE user_release_ratings (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    release_id BIGINT NOT NULL REFERENCES releases(id) ON DELETE CASCADE,
    rating SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 10),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (user_id, release_id)
);

CREATE INDEX idx_user_release_ratings_release_id
    ON user_release_ratings (release_id);