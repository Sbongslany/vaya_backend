-- +goose Up
ALTER TABLE trips ADD COLUMN IF NOT EXISTS cancellation_reason TEXT;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS cancelled_by UUID REFERENCES auth.users(id);
ALTER TABLE trips ADD COLUMN IF NOT EXISTS cancelled_at TIMESTAMPTZ;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS cancellation_fee NUMERIC(10,2);

-- +goose Down
ALTER TABLE trips DROP COLUMN IF EXISTS cancellation_fee;
ALTER TABLE trips DROP COLUMN IF EXISTS cancelled_at;
ALTER TABLE trips DROP COLUMN IF EXISTS cancelled_by;
ALTER TABLE trips DROP COLUMN IF EXISTS cancellation_reason;