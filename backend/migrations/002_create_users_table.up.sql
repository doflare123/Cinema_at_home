CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    telegram_id BIGINT UNIQUE,
    telegram_username VARCHAR(255),
    display_name VARCHAR(255) NOT NULL DEFAULT '',
    avatar_url TEXT,
    username VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    role_id INT REFERENCES roles(id) DEFAULT 1,
    status VARCHAR(16) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'active', 'rejected', 'blocked'))
);
