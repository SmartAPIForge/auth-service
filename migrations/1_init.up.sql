CREATE TABLE IF NOT EXISTS users
(
    id       INTEGER PRIMARY KEY,
    email    TEXT NOT NULL UNIQUE,
    password BLOB NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_email ON users (email);