CREATE TABLE IF NOT EXISTS films (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    small_description VARCHAR(100) NOT NULL,
    duration INT NOT NULL,
    release_date DATE NOT NULL,
    country VARCHAR(255) NOT NULL,
    poster VARCHAR(255) NOT NULL,
    rating_kp FLOAT NOT NULL
)