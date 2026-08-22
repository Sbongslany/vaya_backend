-- +goose Up
ALTER TABLE trips ADD COLUMN IF NOT EXISTS surge_multiplier NUMERIC(4,2) DEFAULT 1.00;

-- +goose Down
ALTER TABLE trips DROP COLUMN IF EXISTS surge_multiplier;