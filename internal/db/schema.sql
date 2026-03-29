CREATE TABLE users (
    id         UUID        PRIMARY KEY,
    email      TEXT        NOT NULL UNIQUE,
    provider   TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
