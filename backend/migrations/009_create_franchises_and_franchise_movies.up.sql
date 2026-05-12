CREATE TABLE IF NOT EXISTS franchises (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS franchise_movies (
    id SERIAL PRIMARY KEY,
    franchise_id INT NOT NULL REFERENCES franchises(id) ON DELETE CASCADE,
    film_id INT NOT NULL REFERENCES films(id) ON DELETE CASCADE,
    part_number INT NOT NULL,
    release_order INT NOT NULL,
    chronology_order INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT franchise_movies_part_number_positive CHECK (part_number > 0),
    CONSTRAINT franchise_movies_release_order_positive CHECK (release_order > 0),
    CONSTRAINT franchise_movies_chronology_order_positive CHECK (chronology_order > 0),
    CONSTRAINT franchise_movies_franchise_film_unique UNIQUE (franchise_id, film_id)
);

CREATE INDEX IF NOT EXISTS idx_franchise_movies_film_id
    ON franchise_movies(film_id);
