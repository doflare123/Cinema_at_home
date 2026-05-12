CREATE TABLE IF NOT EXISTS expectation_votes (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_type VARCHAR(16) NOT NULL,
    film_id INT REFERENCES films(id) ON DELETE CASCADE,
    franchise_id INT REFERENCES franchises(id) ON DELETE CASCADE,
    target_id INT GENERATED ALWAYS AS (
        CASE
            WHEN target_type = 'movie' THEN film_id
            WHEN target_type = 'franchise' THEN franchise_id
            ELSE NULL
        END
    ) STORED,
    vote_type VARCHAR(16) NOT NULL,
    score SMALLINT,
    comment TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT expectation_votes_target_type_check CHECK (target_type IN ('movie', 'franchise')),
    CONSTRAINT expectation_votes_vote_type_check CHECK (vote_type IN ('score', 'refuse')),
    CONSTRAINT expectation_votes_target_check CHECK (
        (target_type = 'movie' AND film_id IS NOT NULL AND franchise_id IS NULL)
        OR
        (target_type = 'franchise' AND franchise_id IS NOT NULL AND film_id IS NULL)
    ),
    CONSTRAINT expectation_votes_score_check CHECK (
        (vote_type = 'score' AND score BETWEEN 1 AND 10)
        OR
        (vote_type = 'refuse' AND score IS NULL)
    ),
    CONSTRAINT expectation_votes_user_target_unique UNIQUE (user_id, target_type, target_id)
);

CREATE INDEX IF NOT EXISTS idx_expectation_votes_target
    ON expectation_votes(target_type, target_id);

CREATE INDEX IF NOT EXISTS idx_expectation_votes_target_score
    ON expectation_votes(target_type, target_id)
    WHERE vote_type = 'score';
