-- +goose Up
CREATE TABLE users (
    id uuid PRIMARY KEY,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT users_nonzero_id CHECK (id <> '00000000-0000-0000-0000-000000000000')
);

CREATE TABLE conversations (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users(id),
    title text NOT NULL CHECK (char_length(title) BETWEEN 1 AND 200),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT conversations_nonzero_id CHECK (id <> '00000000-0000-0000-0000-000000000000')
);

CREATE INDEX conversations_user_updated_idx
    ON conversations (user_id, updated_at DESC, id DESC);

-- +goose Down
DROP TABLE conversations;
DROP TABLE users;
