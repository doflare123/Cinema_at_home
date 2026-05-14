CREATE TABLE IF NOT EXISTS movie_proposals (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    small_description TEXT NOT NULL,
    duration INT NOT NULL,
    release_date INT NOT NULL,
    country VARCHAR(255) NOT NULL,
    poster TEXT NOT NULL,
    rating_kp NUMERIC(4,2) NOT NULL DEFAULT 0,
    source VARCHAR(32) NOT NULL DEFAULT 'manual',
    status VARCHAR(16) NOT NULL DEFAULT 'pending',
    proposed_by_user_id INT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    moderated_by_user_id INT NULL REFERENCES users(id) ON DELETE RESTRICT,
    moderated_at TIMESTAMP NULL,
    moderation_comment TEXT NOT NULL DEFAULT '',
    film_id INT NULL REFERENCES films(id) ON DELETE RESTRICT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT movie_proposals_source_check CHECK (source IN ('manual', 'kinopoisk')),
    CONSTRAINT movie_proposals_status_check CHECK (status IN ('pending', 'approved', 'rejected')),
    CONSTRAINT movie_proposals_duration_check CHECK (duration > 0),
    CONSTRAINT movie_proposals_release_date_check CHECK (release_date > 0),
    CONSTRAINT movie_proposals_rating_kp_check CHECK (rating_kp >= 0 AND rating_kp <= 10),
    CONSTRAINT movie_proposals_moderation_audit_check CHECK (
        (status = 'pending' AND moderated_by_user_id IS NULL AND moderated_at IS NULL AND film_id IS NULL)
        OR (status = 'approved' AND moderated_by_user_id IS NOT NULL AND moderated_at IS NOT NULL AND film_id IS NOT NULL)
        OR (status = 'rejected' AND moderated_by_user_id IS NOT NULL AND moderated_at IS NOT NULL AND film_id IS NULL)
    )
);

CREATE INDEX IF NOT EXISTS idx_movie_proposals_status_created
    ON movie_proposals(status, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_movie_proposals_proposed_by_created
    ON movie_proposals(proposed_by_user_id, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_movie_proposals_moderated_by_user_id
    ON movie_proposals(moderated_by_user_id);
