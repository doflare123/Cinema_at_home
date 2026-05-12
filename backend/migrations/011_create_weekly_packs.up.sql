CREATE TABLE IF NOT EXISTS weekly_packs (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'draft',
    starts_at TIMESTAMP NULL,
    ends_at TIMESTAMP NULL,
    created_by_user_id INT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT weekly_packs_status_check CHECK (status IN ('draft', 'voting', 'closed', 'archived'))
);

CREATE INDEX IF NOT EXISTS idx_weekly_packs_state
    ON weekly_packs(status);

CREATE TABLE IF NOT EXISTS weekly_pack_movies (
    id SERIAL PRIMARY KEY,
    pack_id INT NOT NULL REFERENCES weekly_packs(id) ON DELETE CASCADE,
    film_id INT NOT NULL REFERENCES films(id) ON DELETE CASCADE,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT weekly_pack_movies_pack_film_unique UNIQUE (pack_id, film_id)
);

CREATE INDEX IF NOT EXISTS idx_weekly_pack_movies_film_id
    ON weekly_pack_movies(film_id);

CREATE TABLE IF NOT EXISTS weekly_pack_votes (
    id SERIAL PRIMARY KEY,
    pack_id INT NOT NULL,
    film_id INT NOT NULL,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vote_value SMALLINT NOT NULL,
    limit_slot SMALLINT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT weekly_pack_votes_pack_movie_fk
        FOREIGN KEY (pack_id, film_id)
        REFERENCES weekly_pack_movies(pack_id, film_id)
        ON DELETE CASCADE,
    CONSTRAINT weekly_pack_votes_vote_value_check CHECK (vote_value IN (3, 2, 1, 0, -2)),
    CONSTRAINT weekly_pack_votes_limit_slot_check CHECK (
        (vote_value = 3 AND limit_slot = 1)
        OR (vote_value = 2 AND limit_slot BETWEEN 1 AND 2)
        OR (vote_value = 1 AND limit_slot BETWEEN 1 AND 2)
        OR (vote_value = -2 AND limit_slot = 1)
        OR (vote_value = 0 AND limit_slot IS NULL)
    ),
    CONSTRAINT weekly_pack_votes_user_movie_unique UNIQUE (pack_id, user_id, film_id)
);

CREATE INDEX IF NOT EXISTS idx_weekly_pack_votes_pack_movie
    ON weekly_pack_votes(pack_id, film_id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_weekly_pack_votes_pack_user_value_slot
    ON weekly_pack_votes(pack_id, user_id, vote_value, limit_slot)
    WHERE vote_value <> 0;
