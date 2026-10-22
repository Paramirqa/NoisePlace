-- Artists

CREATE TABLE artists (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL,
    normalized_name TEXT NOT NULL,
    sort_name TEXT,
    description TEXT,
    image_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT artists_name_not_empty
        CHECK (LENGTH(TRIM(name)) > 0),

    CONSTRAINT artists_normalized_name_not_empty
        CHECK (LENGTH(TRIM(normalized_name)) > 0)
);

CREATE INDEX idx_artists_normalized_name
    ON artists (normalized_name);


-- Releases

CREATE TABLE releases (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title TEXT NOT NULL,
    release_type TEXT NOT NULL,
    release_date DATE,
    description TEXT,
    cover_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT releases_title_not_empty
        CHECK (LENGTH(TRIM(title)) > 0),

    CONSTRAINT releases_release_type_check
        CHECK (
            release_type IN (
                'album',
                'ep',
                'single',
                'compilation',
                'mixtape',
                'live',
                'other'
            )
        )
);

CREATE INDEX idx_releases_title
    ON releases (title);

CREATE INDEX idx_releases_release_date
    ON releases (release_date);


-- Release artists

CREATE TABLE release_artists (
    release_id BIGINT NOT NULL,
    artist_id BIGINT NOT NULL,
    role TEXT NOT NULL DEFAULT 'primary',
    position INTEGER NOT NULL DEFAULT 0,

    PRIMARY KEY (release_id, artist_id, role),

    CONSTRAINT release_artists_release_fk
        FOREIGN KEY (release_id)
        REFERENCES releases (id)
        ON DELETE CASCADE,

    CONSTRAINT release_artists_artist_fk
        FOREIGN KEY (artist_id)
        REFERENCES artists (id)
        ON DELETE RESTRICT,

    CONSTRAINT release_artists_role_check
        CHECK (
            role IN (
                'primary',
                'featured',
                'producer'
            )
        ),

    CONSTRAINT release_artists_position_check
        CHECK (position >= 0)
);

CREATE INDEX idx_release_artists_artist_release
    ON release_artists (artist_id, release_id);


-- Artist external IDs

CREATE TABLE artist_external_ids (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    artist_id BIGINT NOT NULL,
    source TEXT NOT NULL,
    external_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT artist_external_ids_artist_fk
        FOREIGN KEY (artist_id)
        REFERENCES artists (id)
        ON DELETE CASCADE,

    CONSTRAINT artist_external_ids_source_not_empty
        CHECK (LENGTH(TRIM(source)) > 0),

    CONSTRAINT artist_external_ids_external_id_not_empty
        CHECK (LENGTH(TRIM(external_id)) > 0),

    CONSTRAINT artist_external_ids_source_external_unique
        UNIQUE (source, external_id),

    CONSTRAINT artist_external_ids_artist_source_unique
        UNIQUE (artist_id, source)
);


-- Release external IDs

CREATE TABLE release_external_ids (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    release_id BIGINT NOT NULL,
    source TEXT NOT NULL,
    external_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT release_external_ids_release_fk
        FOREIGN KEY (release_id)
        REFERENCES releases (id)
        ON DELETE CASCADE,

    CONSTRAINT release_external_ids_source_not_empty
        CHECK (LENGTH(TRIM(source)) > 0),

    CONSTRAINT release_external_ids_external_id_not_empty
        CHECK (LENGTH(TRIM(external_id)) > 0),

    CONSTRAINT release_external_ids_source_external_unique
        UNIQUE (source, external_id),

    CONSTRAINT release_external_ids_release_source_unique
        UNIQUE (release_id, source)
);