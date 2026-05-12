CREATE TABLE IF NOT EXISTS reviews (
    id SERIAL PRIMARY KEY,
    film_id INT NOT NULL REFERENCES films(id) ON DELETE CASCADE,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    mode VARCHAR(16) NOT NULL,
    score SMALLINT NULL,
    final_score NUMERIC(4,2) NOT NULL,
    criteria_scores JSONB NOT NULL DEFAULT '{}'::jsonb,
    comment TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT reviews_mode_check CHECK (mode IN ('simple', 'criteria')),
    CONSTRAINT reviews_score_check CHECK (
        (mode = 'simple' AND score BETWEEN 1 AND 10)
        OR (mode = 'criteria' AND score IS NULL)
    ),
    CONSTRAINT reviews_final_score_check CHECK (final_score >= 1 AND final_score <= 10),
    CONSTRAINT reviews_user_film_unique UNIQUE (film_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_reviews_film_id
    ON reviews(film_id);

CREATE INDEX IF NOT EXISTS idx_reviews_user_id
    ON reviews(user_id);

CREATE OR REPLACE FUNCTION set_review_final_score()
RETURNS TRIGGER AS $$
DECLARE
    item RECORD;
    score_value INT;
    score_text TEXT;
    total NUMERIC := 0;
    score_count INT := 0;
BEGIN
    IF NEW.mode = 'simple' THEN
        IF NEW.score IS NULL OR NEW.score < 1 OR NEW.score > 10 THEN
            RAISE EXCEPTION 'simple review score must be between 1 and 10';
        END IF;
        NEW.criteria_scores := '{}'::jsonb;
        NEW.final_score := NEW.score;
    ELSIF NEW.mode = 'criteria' THEN
        NEW.score := NULL;
        IF NEW.criteria_scores IS NULL
            OR jsonb_typeof(NEW.criteria_scores) <> 'object'
            OR NEW.criteria_scores = '{}'::jsonb THEN
            RAISE EXCEPTION 'criteria_scores are required for criteria reviews';
        END IF;

        FOR item IN SELECT key, value FROM jsonb_each(NEW.criteria_scores)
        LOOP
            score_text := item.value #>> '{}';
            IF btrim(item.key) = ''
                OR jsonb_typeof(item.value) <> 'number'
                OR score_text !~ '^[0-9]+$' THEN
                RAISE EXCEPTION 'criteria review scores must be integer values';
            END IF;

            score_value := score_text::INT;
            IF score_value < 1 OR score_value > 10 THEN
                RAISE EXCEPTION 'criteria review scores must be between 1 and 10';
            END IF;

            total := total + score_value;
            score_count := score_count + 1;
        END LOOP;

        NEW.final_score := ROUND(total / score_count, 2);
    ELSE
        RAISE EXCEPTION 'invalid review mode';
    END IF;

    IF TG_OP = 'INSERT' AND NEW.created_at IS NULL THEN
        NEW.created_at := CURRENT_TIMESTAMP;
    END IF;
    NEW.updated_at := CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS reviews_set_final_score ON reviews;
CREATE TRIGGER reviews_set_final_score
BEFORE INSERT OR UPDATE OF mode, score, final_score, criteria_scores, comment
ON reviews
FOR EACH ROW
EXECUTE FUNCTION set_review_final_score();
