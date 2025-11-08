CREATE TABLE IF NOT EXISTS films (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL UNIQUE,
    description TEXT NOT NULL,
    small_description VARCHAR(255) NOT NULL,
    duration INT NOT NULL,
    release_date INT NOT NULL,
    country VARCHAR(255) NOT NULL,
    poster VARCHAR(255) NOT NULL,
    rating_kp FLOAT NOT NULL
)