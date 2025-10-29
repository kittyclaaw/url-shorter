-- +goose Up
CREATE TABLE urls(
     id SERIAL PRIMARY KEY,
     original_url TEXT NOT NULL CHECK (original_url LIKE 'http%'),
     short_code TEXT UNIQUE NOT NULL,
     created_at TIMESTAMP DEFAULT now(),
     updated_at TIMESTAMP DEFAULT now(),
     click_count INT DEFAULT 0
);

-- +goose Down
DROP TABLE urls CASCADE;