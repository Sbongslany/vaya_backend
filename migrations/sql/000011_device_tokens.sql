-- +goose Up
CREATE TABLE IF NOT EXISTS device_tokens (
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    token TEXT NOT NULL,
    device_type VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, token)
);

-- +goose Down
DROP TABLE IF EXISTS device_tokens;