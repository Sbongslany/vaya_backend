-- +goose Up
ALTER TABLE trips ADD COLUMN start_pin VARCHAR(4);

-- +goose Down
ALTER TABLE trips DROP COLUMN IF EXISTS start_pin;