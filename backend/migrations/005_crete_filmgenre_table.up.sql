CREATE TABLE IF NOT EXISTS film_genre (
    film_id INT REFERENCES films(id),
    genre_id INT REFERENCES genres(id),
    PRIMARY KEY (film_id, genre_id)
);