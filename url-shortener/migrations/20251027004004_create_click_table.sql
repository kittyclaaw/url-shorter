-- +goose Up
CREATE TABLE clicks(
   id SERIAL PRIMARY KEY,
   url_id INTEGER REFERENCES urls(id),
   ip_address TEXT,
   user_agent TEXT,
   referer TEXT,
   created_at TIMESTAMP DEFAULT now()
);

-- +goose Down
DROP TABLE clicks;